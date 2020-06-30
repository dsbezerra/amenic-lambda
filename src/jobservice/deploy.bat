echo off

REM Set env variables
set GOOS=linux
set GOARCH=amd64
set CGO_ENABLED=0

REM Cleanup
echo Cleaning up...

del bin\main 2>nul
del bin\worker 2>nul
del main.zip 2>nul
del worker.zip 2>nul

echo Cleaning up... Finished.

REM Build schedule jobs and worker binaries
echo Building...
echo Building main...

go build -o bin/main

echo Building main... Finished.

cd worker
echo Building worker...

go build -o ../bin/worker

echo Building worker... Finished.
echo Building... Finished.

REM Zip both binaries
echo Zipping binaries...

cd ..
%USERPROFILE%\Go\bin\build-lambda-zip.exe -o main.zip bin\main
%USERPROFILE%\Go\bin\build-lambda-zip.exe -o worker.zip bin\worker

echo Zipping binaries... Finished.

REM Package and deploy to amenic-jobs stack
echo Packaging and deploying...

aws cloudformation package --template-file template.yaml --output-template-file output-template.yaml --s3-bucket amenic-zips
aws cloudformation deploy --template-file output-template.yaml --stack-name amenic-jobs --capabilities CAPABILITY_IAM

echo Packaging and deploying... Finished.
