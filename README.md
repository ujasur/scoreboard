## Scoreboard
This is small app that I wrote for work/story pointing during scrum planning by teams, in one of my old workplaces. Voters/team members decide complexity of the task under discussion by selecting points on the app which calculates overall score of the task. 

### What is needed to run.
* Go >= 1.13
* Node.js
* Npx and npm for development purposes

### Config.
Teams and their preferences should be put at `./config/teams.json`. For example, look at  `./config/teams.example.json`.

### Dev running.
* `./dev_setup.sh` - run once to install necessary tools to compile jsx files.
* `./dev_run.sh` - it builds go app and transforms jsx to js. 
* Serves app from the port specified in configuration.

### Bundle everything.
* `make` - it compiles backend and ui, puts all necessary asset files into `artifact` folder.

![Alt text](./screenshot.jpg?raw=true "Example")
