# phace

Reverse engineering OSX/macOS `*.photoslibrary` format. Primary goal is to
export face tagging data from the embedded databases.

## Status

This project is very much a work in progress. Additionally, Apple appears to have made significant changes to the `*.photoslibrary` structure in macOS 10.15 Catalina (maybe 10.14?). Phace has not been updated for these newer versions.

```sh
$ sqlite3 ./database/photos.db .dump
```

```sql
PRAGMA foreign_keys=OFF;
BEGIN TRANSACTION;
CREATE TABLE LiGlobals (modelId integer primary key, keyPath varchar, value varchar, blobValue blob);
INSERT INTO LiGlobals VALUES(1,'metaSchemaVersion','3',NULL);
INSERT INTO LiGlobals VALUES(2,'libraryVersion','6000',NULL);
INSERT INTO LiGlobals VALUES(3,'libraryCompatibleBackToVersion','6000',NULL);
CREATE TABLE LiLibHistory (modelId integer primary key, modDate timestamp, eventType varchar, metaSchemaVersion integer, libraryVersion integer, comment varchar);
INSERT INTO LiLibHistory VALUES(1,466369847.60223102568,'deprecate',3,6000,'metaSchema.db and photos.db replaced by Photos.sqlite');
COMMIT;
```

Specifically that last comment: `metaSchema.db and photos.db replaced by Photos.sqlite`
