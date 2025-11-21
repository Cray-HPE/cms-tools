#!/bin/bash
#
# MIT License
#
# (C) Copyright 2025 Hewlett Packard Enterprise Development LP
#
# Permission is hereby granted, free of charge, to any person obtaining a
# copy of this software and associated documentation files (the "Software"),
# to deal in the Software without restriction, including without limitation
# the rights to use, copy, modify, merge, publish, distribute, sublicense,
# and/or sell copies of the Software, and to permit persons to whom the
# Software is furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included
# in all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
# THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
# OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
# ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
# OTHER DEALINGS IN THE SOFTWARE.
#

#
# convert_cmsdev_logs.sh
# Manual execution (for testing or troubleshooting):
# Use default directory
# /usr/local/bin/convert_cmsdev_logs.sh

# Use custom directory
# /usr/local/bin/convert_cmsdev_logs.sh /path/to/custom/logdir
#
# Convert existing cmsdev log files and artifacts to new timestamped directory convention
#

# Default log directory
CMSDEV_LOGDIR="/opt/cray/tests/install/logs/cmsdev"

# Allow override via command line argument
if [ $# -eq 1 ]; then
    CMSDEV_LOGDIR="$1"
fi

echo "Converting existing cmsdev logs to new timestamped directory format..."
echo "Log directory: $CMSDEV_LOGDIR"

# Check if cmsdev.log exists
if [ ! -f "$CMSDEV_LOGDIR/cmsdev.log" ]; then
    echo "No existing cmsdev.log file found at $CMSDEV_LOGDIR/cmsdev.log"
    exit 0
fi

# Extract unique main run tags (format: run=GGqaQ, excluding subtags like run=GGqaQ-cfs)
unique_tags=$(grep -oE 'run=[a-zA-Z0-9]{5}' "$CMSDEV_LOGDIR/cmsdev.log" 2>/dev/null | grep -vE 'run=[a-zA-Z0-9]{5}-[a-z0-9]+' 2>/dev/null | sort -u | cut -d'=' -f2)

if [ -z "$unique_tags" ]; then
    echo "No run tags found in cmsdev.log, keeping original file"
    exit 0
fi

echo "Found run tags: $(echo $unique_tags | tr '\n' ' ')"

for tag in $unique_tags; do
    echo "Processing run tag: $tag"
    
    # Find the earliest timestamp for this run tag
    earliest_time=$(grep -E " run=${tag}( |-[a-z0-9]+ )" "$CMSDEV_LOGDIR/cmsdev.log" 2>/dev/null | head -1 | cut -d'"' -f2)
    
    echo "Earliest timestamp: $earliest_time"
    
    if [ -z "$earliest_time" ]; then
        echo "No timestamp found for tag $tag, skipping..."
        continue
    fi
    
    # Convert timestamp to YYMMDD_HHMMSS_microseconds format
    # Input format examples: 2020-05-21T14:19:32-05:00 or 2025-11-21T17:06:19.842986851Z
    if echo "$earliest_time" | grep -qE '[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}'; then
        # Extract year, month, day, hour, minute, second, and microseconds
        year=$(echo "$earliest_time" | cut -d'T' -f1 | cut -d'-' -f1 | cut -c3-4)
        month=$(echo "$earliest_time" | cut -d'T' -f1 | cut -d'-' -f2)
        day=$(echo "$earliest_time" | cut -d'T' -f1 | cut -d'-' -f3)
        time_part=$(echo "$earliest_time" | cut -d'T' -f2)
        hour=$(echo "$time_part" | cut -d':' -f1)
        minute=$(echo "$time_part" | cut -d':' -f2)
        second_part=$(echo "$time_part" | cut -d':' -f3)
        
        # Extract seconds and microseconds, handle timezone/Z suffix
        second=$(echo "$second_part" | sed -E 's/([0-9]{2})(\.[0-9]+)?[Z+-].*/\1/')
        microsec_part=$(echo "$second_part" | grep -oE '\.[0-9]+' | tr -d '.' || echo "")
        
        # Pad or truncate microseconds to 6 digits
        if [ -n "$microsec_part" ]; then
            # Pad with zeros to 6 digits or truncate to 6 digits
            microsec=$(printf "%-6s" "$microsec_part" | tr ' ' '0' | cut -c1-6)
        else
            microsec="000000"
        fi
        
        timestamp_dir="${year}${month}${day}_${hour}${minute}${second}_${microsec}"
    else
        echo "Unable to parse timestamp format for $earliest_time, using fallback"
        timestamp_dir=$(echo "$earliest_time" | tr -cd '0-9' | head -c 17)
    fi
    
    echo "Target directory: $timestamp_dir"
    
    # Create timestamped directory if it doesn't exist
    target_dir="$CMSDEV_LOGDIR/$timestamp_dir"
    if [ -d "$target_dir" ]; then
        echo "Directory $target_dir already exists, skipping tag $tag"
        continue
    fi
    
    mkdir -p "$target_dir"
    if [ $? -eq 0 ]; then
        echo "Created directory: $target_dir"
    else
        echo "Failed to create directory: $target_dir, skipping tag $tag"
        continue
    fi
    
    # Extract logs for this run tag and strip run= tags
    grep -E " run=${tag}( |-[a-z0-9]+ )" "$CMSDEV_LOGDIR/cmsdev.log" 2>/dev/null | sed -E "s/ run=${tag}( |-[a-z0-9]+ )/ /" > "$target_dir/cmsdev.log"
    if [ $? -eq 0 ]; then
        log_count=$(wc -l < "$target_dir/cmsdev.log" 2>/dev/null || echo "0")
        echo "Extracted $log_count log entries for run $tag to $target_dir/cmsdev.log"
    else
        echo "Failed to extract logs for tag $tag"
    fi
    
    # Check if there's a "No artifacts saved" message for this run tag
    no_artifacts_msg=$(grep -E " run=${tag}( |-[a-z0-9]+ )" "$CMSDEV_LOGDIR/cmsdev.log" 2>/dev/null | grep -E 'msg="No artifacts saved\. Removing empty artifact directory:' 2>/dev/null || echo "")
    
    if [ -n "$no_artifacts_msg" ]; then
        echo "Found 'No artifacts saved' message for run tag $tag, skipping artifact processing"
    else
        # Look for artifact file references in the logs for this run tag
        # Pattern matches: msg="ARTIFACTS environment variable not set. Defaulting to '/opt/cray/tests/install/logs/cmsdev/artifacts-2025-11-21T17:06:19.842986851Z'"
        artifact_file=$(grep -E " run=${tag}( |-[a-z0-9]+ )" "$CMSDEV_LOGDIR/cmsdev.log" 2>/dev/null | grep -oE "'/opt/cray/tests/install/logs/cmsdev/artifacts-[^']*'" 2>/dev/null | head -1 | tr -d "'" || echo "")
        artifact_compressed="${artifact_file}.tgz"

        if [ -n "$artifact_compressed" ]; then
            echo "Found artifact reference: $artifact_compressed"
            
            if [ -f "$artifact_compressed" ]; then
                echo "Moving artifact file $artifact_compressed to $target_dir/"
                mv "$artifact_compressed" "$target_dir/" || echo "Failed to move artifact file $artifact_compressed"
            else
                echo "Artifact file $artifact_compressed referenced but not found"
            fi
        else
            echo "No artifact file reference found for run tag $tag"
        fi
    fi
    
    echo "Completed processing for run tag: $tag"
    echo "--------------------------------"
done

# Remove the single cmsdev.log file after processing all tags
echo "Removing original cmsdev.log file"
rm -f "$CMSDEV_LOGDIR/cmsdev.log"
echo "Log conversion completed successfully"
echo "Processed $(echo "$unique_tags" | wc -w) run tags"