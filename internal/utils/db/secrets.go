package db

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go/aws"
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func initSecretsConfig() *secretsmanager.Client {
	config, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	return secretsmanager.NewFromConfig(config)
}

func retrieveCredentials(secretID string) (string, string) {
	secretUsername := os.Getenv("DB_USERNAME")
	secretPassword := os.Getenv("DB_PASSWORD")
	if secretUsername != "" && secretPassword != "" {
		return secretUsername, secretPassword
	}

	secrets := initSecretsConfig()
	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretID),
		VersionStage: aws.String("AWSCURRENT"), // VersionStage defaults to AWSCURRENT if unspecified
	}

	result, err := secrets.GetSecretValue(context.TODO(), input)
	if err != nil {
		panic(err)
	}
	secretString := []byte(*result.SecretString)

	var secret Credentials
	if err = json.Unmarshal(secretString, &secret); err != nil {
		panic(err)
	}

	return secret.Username, secret.Password
}
