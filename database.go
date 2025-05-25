package main

import (
    "fmt"
    "time"
    "database/sql"
    
    "go.mau.fi/whatsmeow/types"
    waLog "go.mau.fi/whatsmeow/util/log"
)

const (
    CHAT_ACTIVITY_TABLE_NAME = "chat_activity"
)

type RIVAClientDB struct {
    DB  *sql.DB
    Log waLog.Logger
}

func NewRIVAClientDB(db *sql.DB, logger waLog.Logger) *RIVAClientDB {
    cdb := &RIVAClientDB{
        DB:  db,
        Log: logger,
    }

    if err := cdb.SetupTables(); err != nil {
        cdb.Log.Errorf("Failed to initialise database: %v", err)
        panic(err)
    }

    return cdb
}

func (db *RIVAClientDB) SetupTables() error {
    query := fmt.Sprintf(`
    CREATE TABLE IF NOT EXISTS %s (
        chat_jid     TEXT PRIMARY KEY,
        last_message DATETIME NOT NULL
    );`, CHAT_ACTIVITY_TABLE_NAME)

    _, err := db.DB.Exec(query)
    if err != nil {
        db.Log.Errorf("Failed to create %s table: %v", CHAT_ACTIVITY_TABLE_NAME, err)
        return err
    }

    db.Log.Infof("Table %s ensured to exist.", CHAT_ACTIVITY_TABLE_NAME)
    return nil
}

func (db *RIVAClientDB) GetLastInteractionTime(userJID types.JID) (time.Time, bool, error) {
    var timestamp time.Time

    query := fmt.Sprintf(`SELECT last_message FROM %s WHERE chat_jid = ?`, CHAT_ACTIVITY_TABLE_NAME)
    err := db.DB.QueryRow(query, userJID.String()).Scan(&timestamp)
    if err != nil {
        if err == sql.ErrNoRows {
            return time.Time{}, false, nil
        }

        db.Log.Errorf("Failed to query last interaction time for %s: %v", userJID.String(), err)
        return time.Time{}, false, err
    }

    return timestamp, true, nil
}

func (db *RIVAClientDB) UpdateLastInteractionTime(userJID types.JID, timestamp time.Time) error {
    query := fmt.Sprintf(`
    INSERT OR REPLACE INTO %s (
        chat_jid,
        last_message
    ) VALUES (?, ?)`, CHAT_ACTIVITY_TABLE_NAME)

    _, err := db.DB.Exec(query, userJID.String(), timestamp)
    if err != nil {
        db.Log.Errorf("Failed to update last interaction time for %s: %v", userJID.String(), err)
        return err
    }

    return nil
}

