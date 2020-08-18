git add .
git commit -m "save"
git push origin master
set CURRENT_DIR=%cd%
chdir /d %DELPHIPATH%\src\github.com\fpawel\aToolGui
git add .
git commit -m "save"
git push origin master
chdir /d %CURRENT_DIR%