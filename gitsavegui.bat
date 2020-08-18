SET GUIREPO=--git-dir=%DELPHIPATH%\src\github.com\fpawel\aToolGui\.git
git %GUIREPO% add .
git %GUIREPO% commit -m "save"
git %GUIREPO% push origin master