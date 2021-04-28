UPDATE UPDATE UPDATE: See UPDATE in README.md.  This is not a problem
anymore...

I'm running into a problem using package sqlite's backup functions.  I
wrote it up at (hopefully concise-ish) length and with a "minimal" (I
hope!) example program illustrating the behavior at
https://github.com/ratrocket/backupcrash.

Briefly, I'm trying to use both `BackupToDB` (by itself) and also the
group of functions `Backup.BackupInit`, `Backup.Step`, `Backup.Finish`
to make a backup of a database.  I think the problem comes down to
[backup.go, line
74](https://github.com/crawshaw/sqlite/blob/v0.3.2/backup.go#L74) where
the call `C.sqlite3_backup_init(dst.conn, dstCDB, src.conn, srcCDB)`
returns nil and shouldn't.
