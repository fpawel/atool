SET APP_DIR=build
SET GOARCH=386
FOR /F "tokens=*" %%g IN ('git rev-list -1 HEAD') do (SET GIT_COMMIT=%%g)
FOR /F "tokens=*" %%g IN ('powershell -Command "[guid]::NewGuid().ToString()"') do (SET BUILD_UUID=%%g)
FOR /f "tokens=2,3,4,5,6 usebackq delims=:/ " %%a in ('%date% %time%') do echo %%c-%%a-%%b %%d%%e
@echo off
FOR /F "tokens=*" %%g IN ('date /t') do (SET THE_DATE=%%g)
FOR /F "tokens=*" %%g IN ('time /t') do (SET THE_TIME=%%g)
echo %THE_DATE%%THE_TIME%
buildmingw32 go build -o %APP_DIR%\atool.exe ^
-ldflags "-X main.GitCommit=%GIT_COMMIT% -X main.BuildUUID=%BUILD_UUID% -X main.BuildTime=%THE_TIME% -X main.BuildDate=%THE_DATE%" ^
github.com/fpawel/atool/cmd/atool
