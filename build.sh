# This script is used to build the Gold Mine API project

#!/bin/bash

# Create build directory
if [ ! -d "bin" ]; then
    mkdir bin
fi

# Compile source files
go build .

# Move executable to bin directory
mv goldmine-api bin/

echo "stopping service"
sudo systemctl stop goldmine-api.service

echo "copyting binary"
cp bin/goldmine-api /home/sbbs/goldmine

echo "restarting service"
sudo systemctl start goldmine-api

echo "Build completed successfully!"
echo "Executable is located in bin/ directory"
echo "Service has been restarted"