package main

import (
	"database/sql"
	"fmt"
	"time"

	"go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"
)

type RIVAClientDB struct {
    RClient *RIVAClient
    DB      *sql.DB
    Log     waLog.Logger
}

func (*RIVAClientDB) New(rClient *RIVAClient, db *sql.DB, logger waLog.Logger) *RIVAClientDB {
    cdb := &RIVAClientDB{
        RClient: rClient,
        DB:      db,
        Log:     logger,
    }

    if err := cdb.SetupTables(); err != nil {
        cdb.Log.Errorf("Failed to initialise database: %v", err)
        panic(err)
    }

    return cdb
}

func (db *RIVAClientDB) SetupTables() error {
    query := fmt.Sprintf(rBotSqlLastInteractionCreateQuery, rBotSqlLastInteractionTableName)

    _, err := db.DB.Exec(query)
    if err != nil {
        db.Log.Errorf("Failed to create %s table: %v", rBotSqlLastInteractionTableName, err)
        return err
    }

    db.Log.Infof("Table %s ensured to exist.", rBotSqlLastInteractionTableName)
    return nil
}

func (db *RIVAClientDB) GetLastInteractionTime(userJID types.JID) (time.Time, bool, error) {
    var timestamp time.Time

    query := fmt.Sprintf(rBotSqlLastInteractionGetQuery, rBotSqlLastInteractionTableName)
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
    query := fmt.Sprintf(rBotSqlLastInteractionInsertQuery, rBotSqlLastInteractionTableName)

    _, err := db.DB.Exec(query, userJID.String(), timestamp)
    if err != nil {
        db.Log.Errorf("Failed to update last interaction time for %s: %v", userJID.String(), err)
        return err
    }

    return nil
}

