# kobo-to-notion-sync

A simple tool to export your Kobo highlights or annotations to Notion. It uses the Notion API to create a new database page for each highlight or note.

## Export example

![Table view with page details](https://github.com/ldtorres/kobo-to-notion/blob/main/images/notion-table-view.png)

## Prerequisites

Before using this tool, ensure that **NickelMenu** is installed on your Kobo device. Follow the installation instructions here: [NickelMenu Installation](https://pgaskin.net/NickelMenu/).

## Setup Instructions

### 1. Create a Notion Integration Token

- Visit the Notion integrations page: [Notion Integrations](https://www.notion.so/profile/integrations).
- Click on "Create new integration."
- Assign a name to your integration, select the appropriate workspace, ensure it's set to private, and optionally add a logo.
- Once created, copy the "Integration Token" and set it as the `NOTION_TOKEN` environment variable.

### 2. Create a Database in Notion

- Within the selected workspace for the integration, create a new page.
- Inside this page, create a table database by typing `/` and selecting "Table View."
- Name the database as desired.
- Configure the following properties in the table:
  - **Book Title**: Title column (default).
  - **Highlighted Text**: Text.
  - **Annotation**: Text.
  - **Type**: Text.
  - **Date Created**: Date.
  - **Bookmark ID**: Text.

### 3. Link the Integration to the Database

- In the Notion page containing the database, click on the three dots in the upper right corner.
- Select "Connections" and add the integration you created earlier. This allows the integration to access the database.

### 4. Download release

Download the latest release zip file from the [releases page](https://github.com/LDTorres/kobo-to-notion/releases).

## 5. Install tools and Configure on the Kobo

### 5.1 Copy release folder to Kobo

Unzip the folder and copy to the Kobo's `/mnt/onboard/.adds` folder.

Update env file inside of the release folder file with your Notion API key and database ID.

### 5.2 Update Environment Variables

Define the following environment variables on your Kobo device:

```sh
NOTION_TOKEN={replace_with_your_notion_integration_token}
NOTION_DATABASE_ID={replace_with_your_notion_database_id}
KOBO_DB_PATH=/mnt/onboard/.kobo/KoboReader.sqlite
CERT_PATH=/mnt/onboard/.adds/notion_sync/certs/cacert.pem
CREATE_BOOKMARKS_INDIVIDUALLY=true
```

- `NOTION_TOKEN`: The integration token you copied in step 1.
- `NOTION_DATABASE_ID`: The ID of your Notion database, obtainable from the database URL.
- `KOBO_DB_PATH`: Path to the `KoboReader.sqlite` file on your Kobo device.
- `CERT_PATH`: Path to the SSL certificate required for HTTPS connections.
- `CREATE_BOOKMARKS_INDIVIDUALLY`: Set to "true" to create a separate page for each highlight (default behavior), or "false" to group all highlights from the same book on a single page.

### Highlight Organization Options

The tool provides two different ways to organize your highlights in Notion:

1. **Individual Mode** (default): Each highlight is created as a separate page in Notion, with its own properties for book title, highlight text, annotation, etc.

2. **Grouped Mode**: All highlights from the same book are grouped together on a single page, making it easier to review all highlights from a particular book in one place. In this mode, the tool will:
   - Create a page for each book
   - Add all highlights as content blocks in the page
   - If you run the sync multiple times, existing pages will be updated by replacing all blocks with the current set of highlights (preventing duplicates)

To switch between modes, set the `CREATE_BOOKMARKS_INDIVIDUALLY` environment variable to "true" or "false".

### 5.3 Create a Shortcut in NickelMenu

To run the script from your Kobo, you need to add a new menu item in NickelMenu:

1. Open the NickelMenu configuration file (usually found in `/mnt/onboard/.adds/nm/config`).
2. Add the following line:

   ```sh
   menu_item : main : Notion Sync : cmd_spawn : quiet : exec /mnt/onboard/.adds/notion_sync/start.sh

---

## 7. Restart the Kobo

- Safely eject and disconnect your Kobo device from your computer.
- Restart the Kobo to apply the changes.
- A new menu item called **Notion Sync** should now appear in the main menu.

---

## 8. Running the Sync Process

To sync your Kobo highlights with Notion:

1. Navigate to the main menu on your Kobo.
2. Select **Notion Sync** from the NickelMenu.
3. The script will execute and transfer your highlights to Notion.

- Make sure the Kobo is connected to Wifi

---

## Build the Project

### For Local Execution:

Ensure you have Go installed on your system.


```sh
go run main.go
```

OR

```sh
go build -o sync
./sync
```

### For the Kobo Device:

1. Build the Docker image:

   ```sh
   docker build -t arm-go-builder .
   ```

2. Compile the binary for Kobo:

   ```sh
   docker run --rm -it -v $(pwd):/app arm-go-builder bash -c "CC=arm-linux-gnueabi-gcc CGO_ENABLED=1 GOOS=linux GOARCH=arm GOARM=7 go build -trimpath -ldflags '-extldflags -static' -o sync.arm"
   ```

3. Use the build-release.sh script to create a release zip file and copy the content inside the kobo device:

   ```sh
   cd ./scripts
   bash ./build-release.sh
   ```

## VS Code Tasks

This project includes pre-configured VS Code tasks to streamline development. To use them:

1. Open the project in VS Code
2. Press `Ctrl+Shift+P` (or `Cmd+Shift+P` on Mac)
3. Type "Tasks: Run Task" and select one of the following:

### Available Tasks

- **Setup**: Installs all Go dependencies and verifies that everything compiles correctly.
  ```
  go mod download && go mod tidy && go build -v ./...
  ```

- **Test**: Runs all tests in the project with detailed output.
  ```
  go test -v ./...
  ```

- **Build Local**: Compiles the application for local use on your machine.
  ```
  go build -v -o sync
  ```

- **Build Prod**: Performs a complete production build following the steps in the README.
  - Builds the Docker image
  - Compiles the application for Kobo (ARM)
  - Executes the release script
  ```
  docker build ... && docker run ... && bash ./build-release.sh
  ```

These tasks can save time and ensure consistency during development and building processes.

## Remove script from Kobo

1. Remove notion_sync folder from .adds folder 
2. Remove notion_sync.conf from NM config folder

Thats it!

## Troubleshooting

If the synchronization doesn't work as expected:

- **Verify the environment variables**  
  - Ensure the `.env` file is correctly configured with your Notion credentials.

- **Check database path**  
  - Make sure `KOBO_DB_PATH` is set correctly to the `KoboReader.sqlite` file.

- **Confirm NickelMenu configuration**  
  - Ensure that the shortcut to run the script is correctly set up.

- **Check logs of the application**  
  - The logs will be on the following file `/mnt/onboard/.adds/nm/notion_sync/logs/app.log`

---

## Contribution

Feel free to contribute to this project by submitting pull requests or reporting issues. Your feedback is highly appreciated!
