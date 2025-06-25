package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/jmespath/go-jmespath"
	"gopkg.in/yaml.v3"
)

type SecretProviderClass struct {
	Metadata struct {
		Name      string `yaml:"name"`
		Namespace string `yaml:"namespace"`
	} `yaml:"metadata"`
	Spec struct {
		Parameters struct {
			Objects string `yaml:"objects"`
		} `yaml:"parameters"`
	} `yaml:"spec"`
}

type Object struct {
	ObjectName string `yaml:"objectName"`
	ObjectType string `yaml:"objectType"`
	JMESPath   []struct {
		Path        string `yaml:"path"`
		ObjectAlias string `yaml:"objectAlias"`
	} `yaml:"jmesPath"`
}

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("Usage: %s <SecretProviderClass YAML> <output secret.yaml>", os.Args[0])
	}

	filePath := os.Args[1]
	outputPath := os.Args[2]

	yamlData, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	var spc SecretProviderClass
	if err := yaml.Unmarshal(yamlData, &spc); err != nil {
		log.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	var objects []Object
	if err := yaml.Unmarshal([]byte(spc.Spec.Parameters.Objects), &objects); err != nil {
		log.Fatalf("Failed to parse objects: %v", err)
	}

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	client := secretsmanager.NewFromConfig(cfg)
	secretData := make(map[string]string)

	for _, obj := range objects {
		if obj.ObjectType != "secretsmanager" {
			continue
		}

		resp, err := client.GetSecretValue(context.Background(), &secretsmanager.GetSecretValueInput{
			SecretId: &obj.ObjectName,
		})
		if err != nil {
			log.Fatalf("Failed to get secret %s: %v", obj.ObjectName, err)
		}

		var jsonMap map[string]interface{}
		if err := json.Unmarshal([]byte(*resp.SecretString), &jsonMap); err != nil {
			log.Fatalf("Invalid JSON in secret %s: %v", obj.ObjectName, err)
		}

		for _, item := range obj.JMESPath {
			value, err := jmespath.Search(item.Path, jsonMap)
			if err != nil {
				log.Fatalf("JMESPath error for %s: %v", item.Path, err)
			}
			strVal, ok := value.(string)
			if !ok {
				log.Fatalf("Expected string at path %s, got %T", item.Path, value)
			}
			secretData[item.ObjectAlias] = base64.StdEncoding.EncodeToString([]byte(strVal))
		}
	}

	k8sSecret := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Secret",
		"metadata": map[string]interface{}{
			"name":      spc.Metadata.Name,
			"namespace": spc.Metadata.Namespace,
		},
		"type": "Opaque",
		"data": secretData,
	}

	outYaml, err := yaml.Marshal(k8sSecret)
	if err != nil {
		log.Fatalf("Failed to marshal output YAML: %v", err)
	}

	if err := os.WriteFile(outputPath, outYaml, 0644); err != nil {
		log.Fatalf("Failed to write file: %v", err)
	}

	fmt.Printf("✅ Secret YAML created at %s\n", outputPath)
}
