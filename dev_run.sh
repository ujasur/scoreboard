#!/bin/bash
set -e

npx babel --presets @babel/preset-react static/session.jsx --out-file static/session.js
go build . && ./scoreboard
