git add .
git commit -m "save"
git push origin master
SET GUIREPO = --git-dir=%DELPHIPATH%\src\github.com\fpawel\aToolGui\.git
git GUIREPO commit -m "save"
git GUIREPO push origin master