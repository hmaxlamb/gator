package config

import (
    "os"
    "encoding/json"
)

type Config struct {
    Db_url string
    Username string
}

func Read() (Config, error) {
    config_file, err := os.Open("../../.gatorconfig.json")
    if err != nil {
        return Config{}, err
    }

    data := make([]byte, 100)
    _ , err = config_file.Read(data)
    if err != nil {
        return Config{}, err
    }

    var config Config
    err = json.Unmarshal(data, &config)
    if err != nil {
        return Config{}, err
    }

    return config, nil
}
