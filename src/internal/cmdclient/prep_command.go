package cmdclient

import "encoding/json"

type commandConfig interface {
  PrepCommand() [] byte
}

type UpCommand struct {
  Name string `json:"name"`
  File string `json:"file"`
  Schedule int `json:"schedule"`
}

func (uc *UpCommand) PrepCommand() []byte {
  if uc.Name == "" {
    zaplog.Fatalln("No action passed")
  }
  cmdJson, err := json.Marshal(uc)
  if err != nil {
    zaplog.Fatalf("Failed to parsed command: %q", err.Error())
  }
  zaplog.Infoln(string(cmdJson))
  return cmdJson
}


type DownCommand struct {
  Name string `json:"name"`
  Schedule int `json:"schedule"`
}

func (dc *DownCommand) PrepCommand() []byte {
  if dc.Name == "" {
    zaplog.Fatalln("No action passed")
  }
  cmdJson, err := json.Marshal(dc)
  if err != nil {
    zaplog.Fatalf("Failed to parsed command: %q", err.Error())
  }
  zaplog.Infoln(string(cmdJson))
  return cmdJson
}

type UpdateCommand struct {
  Name string `json:"name"`
  File string `json:"file"`
  Schedule int `json:"schedule"`
}

func (udc *UpdateCommand) PrepCommand() []byte {
  if udc.Name == "" {
    zaplog.Fatalln("No action passed")
  }
  cmdJson, err := json.Marshal(udc)
  if err != nil {
    zaplog.Fatalf("Failed to parsed command: %q", err.Error())
  }
  zaplog.Infoln(string(cmdJson))
  return cmdJson
}
