SET APP_DIR=build
SET GOARCH=386
buildmingw32 go build -o %APP_DIR%\atool.exe github.com/fpawel/atool/cmd/atool
