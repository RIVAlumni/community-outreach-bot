package main

import (
	waLog "go.mau.fi/whatsmeow/util/log"
)

type RIVAClientLog struct {
    MainLog waLog.Logger
    DBLog   waLog.Logger
    WMLog   waLog.Logger
}

func (*RIVAClientLog) New() *RIVAClientLog {
    return &RIVAClientLog{
        MainLog: waLog.Stdout("RIVABot::Log", "INFO", true),
        DBLog:   waLog.Stdout("Sqlite3::Log", "ERROR", true),
        WMLog:   waLog.Stdout("WhatsMeow::Log", "INFO", true),
    }
}

