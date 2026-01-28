pipeline {
    agent { label 'terraform_1_10_5' }

    triggers {
        // Trigger on push/merge to master branch via GitHub webhook
        // Requires GitHub plugin to be installed in Jenkins
        // Configure webhook in GitHub: Settings > Webhooks > Add webhook
        // Payload URL: http://your-jenkins-url/github-webhook/
        // Content type: application/json
        // Events: Just the push event
        // Active: checked
        githubPush()
    }

    options {
        // Only run on master branch
        skipDefaultCheckout(false)
    }

    environment {
        GITHUB_TOKEN = credentials('github_token_for_terraform')
        GPG_PASSPHRASE = credentials('GPG-passphrase-for-terraform')
    }

    stages {
        stage('Check Branch') {
            steps {
                script {
                    // Get the branch name from environment or git
                    env.BRANCH_NAME = env.GIT_BRANCH ?: sh(
                        script: 'git rev-parse --abbrev-ref HEAD',
                        returnStdout: true
                    ).trim()
                    
                    // Remove 'origin/' prefix if present
                    env.BRANCH_NAME = env.BRANCH_NAME.replaceAll('origin/', '')
                    
                    echo "Current branch: ${env.BRANCH_NAME}"
                    
                    // Only proceed if on master branch
                    if (env.BRANCH_NAME != 'master') {
                        echo "Skipping pipeline - not on master branch (current: ${env.BRANCH_NAME})"
                        currentBuild.result = 'ABORTED'
                        return
                    }
                    
                    echo "✓ Running on master branch"
                }
            }
        }

        stage('Extract Version') {
            steps {
                script {
                    // Extract version from main.go (looking for: version string = "X.Y.Z")
                    env.VERSION = sh(
                        script: '''
                            grep 'version string = "' main.go | sed 's/.*version string = "\\([^"]*\\)".*/\\1/' | head -1
                        ''',
                        returnStdout: true
                    ).trim()
                    
                    if (!env.VERSION) {
                        error('Could not extract version from main.go. Please ensure version is set in main.go file.')
                    }
                    
                    // Validate version format (X.Y.Z)
                    if (!env.VERSION.matches(/^\\d+\\.\\d+\\.\\d+$/)) {
                        error("Invalid version format in main.go: ${env.VERSION}. Expected format: X.Y.Z")
                    }
                    
                    // Create tag version with 'v' prefix
                    env.VERSION_TAG = "v${env.VERSION}"
                    
                    echo "✓ Extracted version from main.go: ${env.VERSION}"
                    echo "✓ Version tag will be: ${env.VERSION_TAG}"
                }
            }
        }

        stage('Check Prerequisites') {
            steps {
                script {
                    echo 'Checking and installing prerequisites...'
                    
                    // Check if git working tree is clean
                    sh '''
                        if ! git diff-index --quiet HEAD --; then
                            echo "Error: Working tree is not clean. Please commit or stash changes."
                            exit 1
                        fi
                        echo "✓ Git working tree is clean"
                    '''

                    // Install GoReleaser if not installed
                    sh '''
                        if ! command -v goreleaser &> /dev/null; then
                            echo "GoReleaser is not installed. Installing..."
                            # Detect OS and install accordingly
                            if [ -f /etc/os-release ]; then
                                . /etc/os-release
                                OS=$ID
                            else
                                OS=$(uname -s | tr '[:upper:]' '[:lower:]')
                            fi
                            
                            # Install GoReleaser using the official method
                            if command -v curl &> /dev/null; then
                                curl -sL https://github.com/goreleaser/goreleaser/releases/latest/download/goreleaser_Linux_x86_64.tar.gz | sudo tar -xz -C /usr/local/bin goreleaser
                                sudo chmod +x /usr/local/bin/goreleaser
                                echo "✓ GoReleaser installed successfully"
                            else
                                echo "Error: curl is required to install GoReleaser"
                                exit 1
                            fi
                        else
                            echo "✓ GoReleaser is already installed"
                        fi
                    '''

                    // Install GPG if not installed
                    sh '''
                        if ! command -v gpg &> /dev/null; then
                            echo "GPG is not installed. Installing..."
                            # Detect OS and install accordingly
                            if [ -f /etc/os-release ]; then
                                . /etc/os-release
                                OS=$ID
                            else
                                OS=$(uname -s | tr '[:upper:]' '[:lower:]')
                            fi
                            
                            # Install GPG based on OS
                            if [ "$OS" = "ubuntu" ] || [ "$OS" = "debian" ]; then
                                sudo apt-get update -qq
                                sudo apt-get install -y -qq gnupg2
                            elif [ "$OS" = "centos" ] || [ "$OS" = "rhel" ] || [ "$OS" = "fedora" ]; then
                                if command -v dnf &> /dev/null; then
                                    sudo dnf install -y -q gnupg2
                                else
                                    sudo yum install -y -q gnupg2
                                fi
                            elif [ "$OS" = "darwin" ] || [ "$OS" = "macos" ]; then
                                if command -v brew &> /dev/null; then
                                    brew install gnupg
                                else
                                    echo "Error: Homebrew is required to install GPG on macOS"
                                    exit 1
                                fi
                            else
                                echo "Warning: Unknown OS. Attempting to install GPG with generic package manager"
                                if command -v apt-get &> /dev/null; then
                                    sudo apt-get update -qq && sudo apt-get install -y -qq gnupg2
                                elif command -v yum &> /dev/null; then
                                    sudo yum install -y -q gnupg2
                                elif command -v dnf &> /dev/null; then
                                    sudo dnf install -y -q gnupg2
                                else
                                    echo "Error: Could not determine package manager to install GPG"
                                    exit 1
                                fi
                            fi
                            echo "✓ GPG installed successfully"
                        else
                            echo "✓ GPG is already installed"
                        fi
                    '''
                }
            }
        }

        stage('Check Tag Exists') {
            steps {
                script {
                    sh '''
                        if git rev-parse "${VERSION_TAG}" >/dev/null 2>&1; then
                            echo "Error: Tag ${VERSION_TAG} already exists"
                            exit 1
                        else
                            echo "✓ Tag ${VERSION_TAG} does not exist yet"
                        fi
                    '''
                }
            }
        }

        stage('Create Git Tag') {
            steps {
                script {
                    sh '''
                        # Create the tag
                        git tag -a "${VERSION_TAG}" -m "Release ${VERSION_TAG}"
                        echo "✓ Created git tag: ${VERSION_TAG}"
                        
                        # Push the tag to remote
                        git push origin "${VERSION_TAG}"
                        echo "✓ Pushed tag to remote"
                    '''
                }
            }
        }

        stage('Release with GoReleaser') {
            steps {
                script {
                    echo "Starting GoReleaser release for version ${env.VERSION_TAG}..."
                    
                    sh '''
                        # Export environment variables (using single quotes for GPG_PASSPHRASE as per instructions)
                        export GITHUB_TOKEN="${GITHUB_TOKEN}"
                        export GPG_PASSPHRASE='${GPG_PASSPHRASE}'
                        
                        # Run GoReleaser
                        goreleaser release --clean
                        
                        echo "✓ GoReleaser release completed successfully"
                    '''
                }
            }
        }
    }

    post {
        success {
            echo "✓ Release pipeline completed successfully!"
            echo "Version: ${env.VERSION} (tag: ${env.VERSION_TAG})"
            echo "GitHub Release: https://github.com/biot-med/terraform-provider-biot-gen2/releases/tag/${env.VERSION_TAG}"
        }
        failure {
            echo "✗ Release pipeline failed!"
            echo "Please check the logs for details."
        }
        always {
            // Clean up any sensitive data
            sh '''
                unset GITHUB_TOKEN
                unset GPG_PASSPHRASE
            '''
        }
    }
}

