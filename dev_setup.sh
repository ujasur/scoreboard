#!/bin/bash
set -e
npm install --save-dev @babel/core @babel/cli
npm install --save-dev @babel/preset-react
npx babel --presets @babel/preset-react script.js