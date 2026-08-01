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

# Resolve the actual mount point of the device we just created. If a stale
# volume with the same name is already mounted, macOS mounts the new one at
# "/Volumes/<name> 1" (2, 3, ...), so the hardcoded path would point at the
# wrong disk - and cleanup would eject the wrong volume.
RAMDISK_PATH=$(diskutil info "$RAMDISK_DEV" | awk -F': *' '/Mount Point/ {print $2}')
if [ -z "$RAMDISK_PATH" ]; then
  echo "Failed to resolve RAM disk mount point."
  hdiutil detach "$RAMDISK_DEV"
  exit 1
fi
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

# Preserve the run's results before destroying the RAM disk. The processor
# writes WebP files to <work_dir>/<uuid>/output on the RAM disk; copy them
# to the local output folder so ejecting does not discard them.
OUTPUT_DEST="./test/output"
mkdir -p "$OUTPUT_DEST"
find "$RAMDISK_PATH" -type d -name output | while read -r RUN_OUTPUT; do
  echo "Copying results from $RUN_OUTPUT to $OUTPUT_DEST..."
  cp -R "$RUN_OUTPUT"/. "$OUTPUT_DEST"/
done

# Destroy the RAM disk now that the results are safe
echo "Ejecting RAM disk..."
if ! diskutil eject "$RAMDISK_PATH"; then
  echo "diskutil eject failed; detaching device $RAMDISK_DEV..."
  if ! hdiutil detach "$RAMDISK_DEV"; then
    echo "Failed to remove RAM disk $RAMDISK_DEV mounted at $RAMDISK_PATH."
    exit 1
  fi
fi
echo "RAM disk removed."

echo "Done!"

exit 0