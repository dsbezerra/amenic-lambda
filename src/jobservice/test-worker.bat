echo off

REM Set env variables
set GOOS=linux
set GOARCH=amd64
set CGO_ENABLED=0

REM Cleanup
echo Cleaning up...

del bin\worker 2>nul
del worker.zip 2>nul

echo Cleaning up... Finished.

REM Build schedule jobs and worker binaries
echo Building...

cd worker

echo Building worker...

go build -o ..\bin\worker
%USERPROFILE%\Go\bin\build-lambda-zip.exe -o ..\worker.zip ..\bin\worker

echo Building worker... Finished.
echo Building... Finished.

cd ..
sam local invoke AmenicWorker --event events\event-sqs.json
