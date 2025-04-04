# Readme

1. Copy notion_sync.conf to .kobo/.adds/nm/notion_sync.conf
2. Copy notion_sync folder to .kobo/.adds

- Maybe you will need to enable show hidden files in your file explorer.

3. Update .env file inside of notion_sync folder and set the following variables:

NOTION_TOKEN={replace_with_your_notion_integration_token}
NOTION_DATABASE_ID={replace_with_your_notion_database_id}

Safely eject and disconnect your Kobo device from your computer.
Restart the Kobo to apply the changes.
A new menu item called **Notion Sync** should now appear in the Nikel menu.

4. If you want to uninstall the plugin just remove the following files:

.kobo/.adds/nm/notion_sync.conf
.kobo/.adds/nm/notion_sync

Enjoy!!
