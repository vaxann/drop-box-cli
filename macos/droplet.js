// Drop Box CLI droplet: an app to keep in the Dock. Files dropped on it
// open the Raycast "Upload Finder Selection" command with those files;
// launched without files, the command uses the Finder selection.
//
// Built by install.sh: osacompile -l JavaScript -o "Drop Box CLI.app" droplet.js

var COMMAND = "raycast://extensions/vaxann/drop-box-cli/upload-finder-selection";

function openDocuments(docs) {
  launchRaycast(
    docs.map(function (d) {
      return d.toString();
    })
  );
}

function run() {
  launchRaycast([]);
}

function launchRaycast(files) {
  var app = Application.currentApplication();
  app.includeStandardAdditions = true;
  var url = COMMAND;
  if (files.length > 0) {
    url += "?context=" + encodeURIComponent(JSON.stringify({ files: files }));
  }
  app.openLocation(url);
}
