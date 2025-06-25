# spc-to-secret
Read a SecretProviderClass and create a correspondent Secret file using values from AWS Secret Service

### Usage
You need to be authenticated with AWS CLI to use this tool.

```bash 
    go run main.go my-secretproviderclass.yaml output-secret.yaml
```