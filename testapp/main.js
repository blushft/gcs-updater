const { app, BrowserWindow, Menu, protocol, ipcMain } = require("electron");
const log = require("electron-log");
const { autoUpdater } = require("electron-updater");

autoUpdater.logger = log;

autoUpdater.logger.transports.file.level = "info";
log.info("starting");
log.info("now version 0.2.0");

let template = [];
if (process.platform === "darwin") {
  const name = app.getName();
  template.unshift({
    label: name,
    submenu: [
      {
        label: "About " + name,
        role: "about"
      },
      {
        label: "Quit",
        accelerator: "Command+Q",
        click() {
          app.quit();
        }
      }
    ]
  });
}

let win;
function sendStatusToWindow(text) {
  log.info(text);
  win.webContents.send("message", text);
}

function createWindow() {
  win = new BrowserWindow({
    webPreferences: {
      webSecurity: false,
      nodeIntegration: true
    },
    width: 800,
    height: 600
  });
  win.webContents.openDevTools();
  win.on("closed", () => {
    win = null;
  });
  win.loadURL(`file://${__dirname}/index.html#v${app.getVersion()}`);
  return win;
}

autoUpdater.on("checking-for-update", () => {
  sendStatusToWindow("Checking for update...");
});

autoUpdater.on("update-available", info => {
  sendStatusToWindow("Update Available!");
});

autoUpdater.on("update-not-available", info => {
  sendStatusToWindow("No update available :(");
});

autoUpdater.on("error", err => {
  sendStatusToWindow("Error! : " + err);
});

autoUpdater.on("download-progress", progress => {
  let logmsg = "download speed: " + progress.bytesPerSecond;
  logmsg = logmsg + " - Downloaded " + progress.percent + "%";
  logmsg = logmsg + " (" + progress.transferred + "/" + progress.total + ")";
  sendStatusToWindow(logmsg);
});

autoUpdater.on("update-downloaded", (ev, info) => {
  sendStatusToWindow("Finished downloading");

  setTimeout(function() {
    autoUpdater.quitAndInstall();
  }, 5000);
});

app.on("ready", function() {
  const menu = Menu.buildFromTemplate(template);
  Menu.setApplicationMenu(menu);

  createWindow();
});

app.on("window-all-closed", () => {
  app.quit();
});

app.on("ready", function() {
  autoUpdater.checkForUpdatesAndNotify();
});
