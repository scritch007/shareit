ShareMinator
============

Share your files with your friends

Workish

* Browsing
* Removing files
* Creating folder
* Popup for downloading file
* Authentication system (Partial)
* Add User Creation
* Add upload system (Broken for now)
* Add sharing system (Partial)

TODO
* Add Chromecast support for Videos and Music
* Add real database for storing informations

In order to have the polymer theme working you need to install bower.


```bash
   npm install bower
```

then go to the html folder and run bower there it will retrieve all the dependencies for the theme.
   
```bash
   cd $GOPATH/src/github.com/scritch007/shareit/html
   bower update
```

Create a configuration file using the file provided in the examples folder.

then run the following command

```bash
   ShareMinator -c=/PATH_TO_YOUR_CONFIGURATION_FOLDER/config.json
```

<a href="https://godoc.org/github.com/scritch007/shareit"><img src="https://godoc.org/github.com/scritch007/shareit?status.png" alt="GoDoc"></a>
