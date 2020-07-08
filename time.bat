@echo off
FOR /F "tokens=*" %%g IN ('date /t') do (SET THE_DATE=%%g)
FOR /F "tokens=*" %%g IN ('time /t') do (SET THE_TIME=%%g)
echo %THE_DATE%-%THE_TIME%