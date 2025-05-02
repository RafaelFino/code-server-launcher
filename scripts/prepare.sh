#!/bin/bash

args=("$@")
# Check if the script is being run with root privileges
if [ "$EUID" -ne 0 ]; then
    echo "Please run as root"
    exit 1
fi

bash ./base.sh

# Check arg -docker for docker
if [[ " ${args[@]} " =~ " -docker " ]]; then
    echo "Installing docker..."
    bash ./docker.sh
fi

# Check arg -go for golang
if [[ " ${args[@]} " =~ " -go " ]]; then
    echo "Installing golang..."
    bash ./golang.sh
fi

# Check arg -jvm for jvm
if [[ " ${args[@]} " =~ " -jvm " ]]; then
    echo "Installing jvm..."
    bash ./jvm.sh
fi

# Check arg -dotnet for dotnet
if [[ " ${args[@]} " =~ " -dotnet " ]]; then
    echo "Installing dotnet..."
    bash ./dotnet.sh
fi

# Check arg -u for users, splited by comma
if [[ " ${args[@]} " =~ " -u " ]]; then
    echo "Installing user env..."
fi

echo "Installing user env for root..."
sudo -H -u root bash ./user-env.sh root

if [[ " ${args[@]} " =~ " -u " ]]; then
  IFS=',' read -r -a users <<< "${args[@]#*-u }"  
  for user in "${users[@]}"; do 
    # Remove leading and trailing whitespace
    user=$(echo "$user" | sed 's/^[ \t]*//;s/[ \t]*$//')
    echo "Installing user env for $user..."
    
    # Check if user exists
    if ! id "$user" &>/dev/null; then
      echo "User $user does not exist. Creating user..."
        sudo adduser $user
        sudo usermod -aG sudo $user
        # Add user to sudo group
        sudo usermod -aG sudo $user
        # Check if docker group exists
        if ! getent group docker > /dev/null; then
            # Create docker group
            sudo groupadd docker
            # Add user to docker group            
        fi                        
    fi 
  
    sudo -H -u $user bash ./user-env.sh $user
  done
fi