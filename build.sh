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

echo "Build completed successfully!"
echo "Executable is located in bin/ directory"