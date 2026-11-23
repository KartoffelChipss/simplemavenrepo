# Publishing to simplemavenrepo

Publishing requires a token whose user has a **write** grant on the artifact's path. The artifact must also be
registered by an admin first, in the repository it is published to. Unregistered artifacts can't be uploaded or
downloaded, unless the repository has **Register artifacts on publish** turned on: then the first upload of an
artifact registers it, as private or, with **Make them public**, as public. Publishing still needs the write grant.

Every repository is served under its own path, e.g. `http://127.0.0.1:4321/releases`. A new server has one repository,
`releases`, which accepts releases and snapshots. Admins can add more under **Repositories**, for example a
`snapshots` repository that only accepts `-SNAPSHOT` versions. The same artifact can be registered in several
repositories, and permission paths start with the repository, like `releases/com/example/lib`.

Downloading a **private** artifact requires a token with a **read** (or write) grant. Consumers configure the same
`<server>` credentials as below, with an `<id>` matching the repository in their `<repositories>`. Public artifacts
can be downloaded without a token.

For CI, create a token restricted to the artifacts it publishes: choose **Only selected repository paths** when
creating it and add e.g. `releases/com/example/lib` with read & write. A restricted token never gets more access than its user,
and it can only be used for the repository, not for the management API.

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

Point the project at the repository in `pom.xml`. With a separate snapshot repository, use its URL for
`<snapshotRepository>`, e.g. `http://127.0.0.1:4321/snapshots`:

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
            // With a separate snapshot repository:
            // url = uri(if (version.toString().endsWith("SNAPSHOT")) "http://127.0.0.1:4321/snapshots" else "http://127.0.0.1:4321/releases")
            url = uri("http://127.0.0.1:4321/releases")
            isAllowInsecureProtocol = true // only needed for plain http
            credentials(PasswordCredentials::class)
        }
    }
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

Failed uploads explain themselves in the reason phrase, which Maven prints as e.g.
`status code: 403, reason phrase: org.example:lib is not registered in repository snapshots, ...`.

| Status                       | Meaning                                                                                                                                                                                                     |
| ---------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `400 Bad Request`            | The repository doesn't accept this kind of version, e.g. a `-SNAPSHOT` version deployed to a repository that only accepts releases, or there is no repository at that URL.                                  |
| `401 Unauthorized`           | No credentials were sent, the token is wrong or expired, or the username doesn't own the token. Check that the server `<id>` (Maven) or repository `name` (Gradle) matches.                                 |
| `403 Forbidden`              | The artifact isn't registered in this repository, or no grant of the user or their teams covers it, or the token is restricted to other paths. Check `GET /api/me` and the token's access under **Tokens**. |
| `404 Not Found`              | Downloading only: the artifact isn't registered in this repository, or the file doesn't exist. Maven reports this as "Could not find artifact".                                                             |
| `409 Conflict`               | This release version is already published and releases are immutable. Bump the version, or enable "Allow overwriting releases" on the artifact. Snapshots can always be redeployed.                         |
| `415 Unsupported Media Type` | Only `.jar`, `.pom`, `.module`, `.aar`, `.klib`, `.war`, `.ear`, `.zip` and `.tar.gz` files and their `.asc` signatures are stored.                                                                         |
