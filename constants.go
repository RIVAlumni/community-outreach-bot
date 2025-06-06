package main

import (
    "os"
    "gopkg.in/yaml.v2"
)

type RIVAClientConfig struct {
    OrgHeaderFooter  string  `yaml:"org_header_footer"`
    GreetingCooldown float64 `yaml:"greeting_cooldown"`
    GreetingMessage  string  `yaml:"greeting_message"`
}

func GetConf() *RIVAClientConfig {
    config := &RIVAClientConfig{}

    configFile, err := os.ReadFile("config.yaml")
    if err != nil {
        panic(err)
    }

    if err := yaml.Unmarshal(configFile, config); err != nil {
        panic(err)
    }

    return config
}

const (
    rBotSqlFilePath = "./data/rivabot.db"
    rBotSqlLastInteractionTableName   = "chat_activity"
    rBotSqlLastInteractionCreateQuery = `
    CREATE TABLE IF NOT EXISTS %s (
        chat_jid     TEXT PRIMARY KEY,
        last_message DATETIME NOT NULL
    );
    `

    rBotSqlLastInteractionGetQuery    = `
    SELECT last_message FROM %s WHERE chat_jid = ?
    `

    rBotSqlLastInteractionInsertQuery = `
    INSERT OR REPLACE INTO %s (
        chat_jid,
        last_message
    ) VALUES (?, ?)
    `
)

var (
    rBotGreetingCooldownHours   = GetConf().GreetingCooldown
    rBotGreetingMessage = GetConf().GreetingMessage

    rBotOrgHeaderFooter = GetConf().OrgHeaderFooter
)

