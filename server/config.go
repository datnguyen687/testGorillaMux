package server

type config struct {
	Root     string `json:"root"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	PublicIP string `json:"publicIP"`
}
