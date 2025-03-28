#!/bin/bash

# Run build.sh to compile the image-processor binary
./build.sh

# Contributor username
CONTRIBUTOR="iconify"

# Set RAM disk size in MB (e.g., 4096 for 4GB)
RAMDISK_SIZE_MB=4096
RAMDISK_NAME="image-processor-ramdisk"

# Calculate sectors for the RAM disk (1 sector = 512 bytes)
SECTORS=$((RAMDISK_SIZE_MB * 2048))

# Create the RAM disk
RAMDISK_DEV=$(hdiutil attach -nomount ram://$SECTORS | tr -d '[:space:]')
if [ $? -ne 0 ] || [ -z "$RAMDISK_DEV" ]; then
  echo "Failed to create RAM disk."
  exit 1
fi

echo "RAM disk device assigned: $RAMDISK_DEV"

# Format the RAM disk and mount it
echo "Formatting the RAM disk with HFS+..."
diskutil erasevolume HFS+ "$RAMDISK_NAME" "$RAMDISK_DEV"
if [ $? -ne 0 ]; then
  echo "Failed to format RAM disk."
  hdiutil detach "$RAMDISK_DEV"
  exit 1
fi

RAMDISK_PATH="/Volumes/$RAMDISK_NAME"
echo "RAM disk is mounted at: $RAMDISK_PATH"

# Update config.yml for RAM disk paths
CONFIG_FILE="./config.yml"
CONFIG_BACKUP="$CONFIG_FILE.bak"

# Backup the original config
cp "$CONFIG_FILE" "$CONFIG_BACKUP"
if [ $? -ne 0 ]; then
  echo "Failed to create backup of $CONFIG_FILE."
  hdiutil detach "$RAMDISK_DEV"
  exit 1
fi

echo "Updating $CONFIG_FILE for RAM disk paths..."

# Extract the current work_dir value from config.yml or default to $RAMDISK_PATH/work
CURRENT_WORK_DIR=$(grep '^work_dir:' "$CONFIG_FILE" | awk '{print $2}' | tr -d '"')

if [ -z "$CURRENT_WORK_DIR" ]; then
  echo "work_dir not found in $CONFIG_FILE. Defaulting to $RAMDISK_PATH/work."
  CURRENT_WORK_DIR=""  # Default relative value if not found
fi

# Prepend $RAMDISK_PATH to the current work_dir value
NEW_WORK_DIR="$RAMDISK_PATH/${CURRENT_WORK_DIR#./}"

# Update the config.yml file with the new work_dir value
sed -i '' "s|^work_dir:.*|work_dir: \"$NEW_WORK_DIR\"|" "$CONFIG_FILE" || echo "work_dir: \"$NEW_WORK_DIR\"" >> "$CONFIG_FILE"

echo "Updated work_dir in config.yml to: $NEW_WORK_DIR"

# Run the image-processor
IMAGE_PROCESSOR_BINARY="./image-processor"
echo "Running the image processor..."
# "$IMAGE_PROCESSOR_BINARY" "-f $CONFIG_FILE -c $CONTRIBUTOR"
./image-processor -f config.yml -c iconify
if [ $? -ne 0 ]; then
  echo "Image processor encountered an error."
  echo "Restoring original config.yml..."
  mv "$CONFIG_BACKUP" "$CONFIG_FILE"
  hdiutil detach "$RAMDISK_DEV"
  exit 1
fi

echo "Image processor finished. Check log output for details."

# Save the run config for debugging
cp "$CONFIG_FILE" "./config_run.yml"

# Restore the original config.yml
echo "Restoring original config.yml..."
mv "$CONFIG_BACKUP" "$CONFIG_FILE"
if [ $? -ne 0 ]; then
  echo "Failed to restore original config.yml."
  exit 1
fi

echo "Done!"

# Uncomment below to automatically destroy the RAM disk
# echo "Ejecting RAM disk..."
# diskutil eject "$RAMDISK_PATH"
# if [ $? -ne 0 ]; then
#   echo "Failed to eject RAM disk."
#   exit 1
# fi
# echo "RAM disk destroyed."

exit 0