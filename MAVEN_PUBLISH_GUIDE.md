# Publishing to simplemavenrepo

Publishing requires a token whose user has a **write** grant on the artifact's path.

## Maven

Add the server to `~/.m2/settings.xml`. The `<id>` must match the repository id in the `pom.xml`.

```xml
<settings>
  <servers>
    <server>
      <id>smr</id>
      <username>alice</username>
      <password>${env.SMR_TOKEN}</password>
    </server>
  </servers>
</settings>
```

Point the project at the repository in `pom.xml`:

```xml
<distributionManagement>
  <repository>
    <id>smr</id>
    <url>http://127.0.0.1:4321/releases</url>
  </repository>
  <snapshotRepository>
    <id>smr</id>
    <url>http://127.0.0.1:4321/releases</url>
  </snapshotRepository>
</distributionManagement>
```

Then publish:

```bash
SMR_TOKEN=smr_... mvn deploy
```

## Gradle

In `build.gradle.kts`:

```kotlin
plugins {
    `java-library`
    `maven-publish`
}

publishing {
    publications {
        create<MavenPublication>("maven") {
            from(components["java"])
        }
    }
    repositories {
        maven {
            name = "smr"
            url = uri("http://127.0.0.1:4321/releases")
            isAllowInsecureProtocol = true // only needed for plain http
            credentials(PasswordCredentials::class)
        }
    }
}

// The server only stores .jar and .pom files
tasks.withType<GenerateModuleMetadata> {
    enabled = false
}
```

`credentials(PasswordCredentials::class)` reads `smrUsername` and `smrPassword` (named after the repository) from
`~/.gradle/gradle.properties`, or from the `ORG_GRADLE_PROJECT_smrUsername` / `ORG_GRADLE_PROJECT_smrPassword`
environment variables:

```properties
smrUsername=alice
smrPassword=smr_...
```

Then publish:

```bash
gradle publish
```

## Troubleshooting

| Status | Meaning |
|---|---|
| `401 Unauthorized` | No credentials were sent, the token is wrong or expired, or the username doesn't own the token. Check that the server `<id>` (Maven) or repository `name` (Gradle) matches. |
| `403 Forbidden` | The token is valid but no grant of the user or their teams covers the artifact. Check `GET /api/me`. |
| `415 Unsupported Media Type` | Only `.jar` and `.pom` files are stored. For Gradle, disable module metadata as shown above. |
