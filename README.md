GoBlogEngine
============

Description
-----------
Blog software written in Golang and designed to run on [AppEngine Standard](https://cloud.google.com/appengine/docs/standard/). 

The goal is to use as many native AppEngine features as possible to keep things simple, lightweight and cheap. Not being portable is an explicit choice (although the data can always be exported). Sane defaults are used for as much out-of-the-box-ness as possible.

Features
--------

- Multiple post versions
- Post drafting and preview
- Image upload and library
- Categories
- Multiple authors
- Post import and export
- Atom feed

Installation
------------
The following instructions are tested and working on Ubuntu Linux, but there's no reason the program shouldn't run on macOS or Windows.

1. Install the [Go tools](https://golang.org/doc/install)
2. Install the [Google Cloud SDK](https://cloud.google.com/sdk/downloads)
3. Install [NodeJS and NPM](https://nodejs.org/en/download/)
4. Install Gulp globally `sudo npm install -g gulp`
5. Clone the repo `git clone https://www.github.com/initialpuppet/goblogengine.git` `cd goblogengine`
6. Install the Node packages  `npm install`
7. Run the Gulp build `gulp`
8. Get the dependencies `cd main` `go get` `cd ..`
9. [Authorise the dev env](https://github.com/golang/appengine/issues/21) against live Google Cloud services: `gcloud auth application-default login`
10. Run the development webserver `./run.sh`
11. Browse to http://localhost:8080/

Deployment
----------
The Cloud SDK enables [single-command deployment](https://cloud.google.com/sdk/gcloud/reference/app/deploy) assuming that the correct configuration is selected. Review the documentation for full details.

By default all static files under the main package, as well as all required Go files are deployed. Using the [app.yaml](https://cloud.google.com/appengine/docs/standard/go/config/appref) file it is possible to specify files to be ignored for deployment. The supplied configuration ignores the `assets` directory and any Markdown files in addition to the files ignored by default.

To deploy: `./deploy.sh`
