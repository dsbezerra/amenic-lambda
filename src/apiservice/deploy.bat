set GOOS=linux
set GOARCH=amd64
set CGO_ENABLED=0
go build -o main
%USERPROFILE%\Go\bin\build-lambda-zip.exe -o main.zip main
aws cloudformation package --template-file template.yaml --output-template-file output-template.yaml --s3-bucket amenic-zips
aws cloudformation deploy --template-file output-template.yaml --stack-name amenic-api --capabilities CAPABILITY_IAM
del main
del main.zip
