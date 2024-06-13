package main

import (
	"log"
	"os"
	"strconv"
)

func getEnvVar(name string) string {
	variable, present := os.LookupEnv(name)
	if !present {
		log.Fatalf("%s env variable is missing", name)
	}
	return variable
}

func getEnvVarInt(name string) int {
	variable := getEnvVar(name)
	num, err := strconv.Atoi(variable)
	if err != nil {
		log.Fatalf("%s env variable is not int", name)
	}
	return num
}
