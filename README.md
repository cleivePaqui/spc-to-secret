# spc-to-secret
Read a SecretProviderClass and create a corresponding Secret file using values from AWS Secrets Manager.

### Usage
You need to be authenticated with AWS CLI to use this tool.

```bash 
    go install github.com/awslabs/spc-to-secret@latest
```

```bash 
    go spc-to-secret my-secretproviderclass.yaml output-secret.yaml
```