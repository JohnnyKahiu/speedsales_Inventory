package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/JohnnyKahiu/speedsales_inventory/api"
	"github.com/JohnnyKahiu/speedsales_inventory/database"
	"github.com/JohnnyKahiu/speedsales_inventory/pkg/products"
	"github.com/JohnnyKahiu/speedsales_inventory/pkg/variables"
	"github.com/joho/godotenv"
)

type dbConf struct {
	Server   string `json:"server"`
	Port     int    `json:"port"`
	DbName   string `json:"database"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type ConfigFile struct {
	Branch         string   `json:"branch"`
	Listen         string   `json:"listen"`
	Port           string   `json:"port"`
	MasterAddr     string   `json:"master_addr"`
	MirrorAddr     string   `json:"Mirror_addr"`
	MirrorPort     string   `json:"mirror_port"`
	ScServer       string   `json:"sc_server"`
	ScPort         int      `json:"sc_port"`
	ServerID       int64    `json:"server_id"`
	ServerName     string   `json:"server_name"`
	ServerBranches []string `json:"server_branches"`
	StockBranch    string   `json:"stock_branch"`
	RemoteSvrs     []string `json:"remote_servers"`
	MasterUrl      string   `json:"master_url"`
	CompanyName    string   `json:"company_name"`
	EtrSocket      string   `json:"etr_socket"`
	EtrType        string   `json:"etr_type"`
	ProductionDisp bool     `json:"production_disp"`
	IndustryMode   string   `json:"industrimode"`
}

func getRunningIPAddress() string {
	addrs, _ := net.InterfaceAddrs()

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				// os.Stdout.WriteString(ipnet.IP.String() + "\n")
				// fmt.Printf("\n%v", ipnet.IP.String())
				return ipnet.IP.String()
			}
		}
	}
	return "0.0.0.0"
}

func (arg *ConfigFile) readConfFile() error {
	file := variables.Fpath + "/config.json"
	fmt.Println("files =", file)
	jsonFile, err := os.Open(file)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
		return err
	}
	byteValue, _ := io.ReadAll(jsonFile)

	json.Unmarshal(byteValue, &arg)

	return nil
}

func infileCache() error {
	fdbPath := fmt.Sprintf("/Data/%v/%v/", os.Getenv("SERVER_NAME"), os.Getenv("DB_NAME"))
	variables.FDBPath = variables.Fpath + fdbPath
	if _, err := os.Stat(variables.FDBPath); errors.Is(err, os.ErrNotExist) {
		os.MkdirAll(variables.FDBPath, os.ModePerm)
	}

	// create a new folder if not exists
	if _, err := os.Stat(variables.Fpath); os.IsNotExist(err) {
		err := os.MkdirAll(variables.Fpath, os.ModePerm)
		if err != nil {
			log.Printf("error creating directory %s: %v", variables.Fpath, err)
			return err
		}
	}

	err := products.ProdMaster.LoadFromDB()
	if err != nil {
		log.Printf("error failed to open a database connection")
		return err
	}
	return nil
}

func main() {
	log.SetOutput(os.Stderr)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	var err error
	// fetch file path from argument
	var serverType, cache string

	flag.StringVar(&variables.Fpath, "path", "", "local files path")
	flag.StringVar(&cache, "cache", "", "cache")
	flag.StringVar(&serverType, "st", "", "server type")

	// get tls certificates
	certFile := flag.String("certfile", "cert.pem", "certificate PEM file")
	keyFile := flag.String("keyfile", "key.pem", "key PEM file")

	isTLS := flag.String("tls", "true", "enable tls")
	initDB := flag.String("initDB", "false", "init db")

	flag.Parse()

	// enable environment files
	err = godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	fmt.Println("server type = ", serverType)
	fmt.Println("init database = ", initDB)
	if variables.Fpath == "" {
		variables.Fpath, err = os.Getwd()
		if err != nil {
			fmt.Println(err)
		}
	}

	// get configuration files
	configs := ConfigFile{}
	err = configs.readConfFile()
	if err != nil {
		log.Println("failed to get config files.   err = ", err)
	}

	conf := database.DBConf{
		Server: os.Getenv("DB_HOST"),
		Port:   os.Getenv("DB_PORT"),
		DbName: os.Getenv("DB_NAME"),
	}
	// make a postgresql database connection
	database.PgPool, err = conf.NewPgPool()
	if err != nil {
		log.Fatalln("\t failed to connect Postgres Pool.    err =", err)
	}
	defer database.PgPool.Close()

	err = infileCache()

	address := getRunningIPAddress()
	if configs.Listen != "card" {
		address = "0.0.0.0"
	}

	port := os.Getenv("PORT")

	r := api.NewRouter()
	if *isTLS == "true" || *isTLS == "1" {
		fmt.Printf("\thttps://%v:%v\n", address, port)
		srv := &http.Server{
			Addr:    address + ":" + port,
			Handler: r,
			TLSConfig: &tls.Config{
				MinVersion:               tls.VersionTLS13,
				PreferServerCipherSuites: true,
			},
		}
		err = srv.ListenAndServeTLS(*certFile, *keyFile)
		if err != nil {
			log.Fatal("failed to start tls server    err =", err)
		}
	} else {
		fmt.Printf("\thttp://%v:%v\n", address, port)
		// http.ListenAndServeTLS(address+":"+port, "localhost.crt", "localhost.key", r)

		http.ListenAndServe(address+":"+port, r)
	}
}
