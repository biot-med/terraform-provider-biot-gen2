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
        skipDefaultCheckout(true)
    }

    environment {
        GITHUB_TOKEN = credentials('github_token_for_terraform')
        GPG_PASSPHRASE = credentials('GPG-passphrase-for-terraform')
    }

    stages {
        stage('Checkout') {
            steps {
                checkout([
                    $class: 'GitSCM',
                    branches: [[name: '*/master']],
                    userRemoteConfigs: [[
                        url: 'https://github.com/biot-med/terraform-provider-biot-gen2.git',
                        credentialsId: 'github_token_for_terraform'
                    ]]
                ])
            }
        }
        
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
                    // Use ==~ operator for regex matching in Groovy
                    if (!(env.VERSION ==~ /^\d+\.\d+\.\d+$/)) {
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
                            echo "GoReleaser is not installed. Installing to user directory..."
                            
                            # Install to user-writable directory
                            INSTALL_DIR="$HOME/.local/bin"
                            mkdir -p "$INSTALL_DIR"
                            
                            # Add to PATH for current session
                            export PATH="$INSTALL_DIR:$PATH"
                            
                            # Install GoReleaser using the official method
                            if command -v curl &> /dev/null; then
                                curl -sL https://github.com/goreleaser/goreleaser/releases/latest/download/goreleaser_Linux_x86_64.tar.gz | tar -xz -C "$INSTALL_DIR" goreleaser
                                chmod +x "$INSTALL_DIR/goreleaser"
                                echo "✓ GoReleaser installed successfully to $INSTALL_DIR"
                            else
                                echo "Error: curl is required to install GoReleaser"
                                exit 1
                            fi
                        else
                            echo "✓ GoReleaser is already installed"
                        fi
                    '''

                    // Check if GPG is installed (cannot install without sudo)
                    sh '''
                        if ! command -v gpg &> /dev/null; then
                            echo "Warning: GPG is not installed."
                            echo "GPG is required for signing releases. Please ensure GPG is pre-installed on the Jenkins agent."
                            echo "The build will continue but may fail during the release stage if GPG signing is required."
                        else
                            echo "✓ GPG is installed"
                            gpg --version | head -1
                        fi
                    '''
                }
            }
        }

        stage('Check Tag Exists') {
            steps {
                script {
                    sh '''
                        # Fetch tags from remote to ensure we have the latest tag information
                        git fetch --tags origin || echo "Warning: Failed to fetch tags, continuing with local check"
                        
                        # Check if tag exists locally or remotely
                        TAG_EXISTS_LOCAL=false
                        TAG_EXISTS_REMOTE=false
                        
                        # Check local tags
                        if git rev-parse "${VERSION_TAG}" >/dev/null 2>&1; then
                            TAG_EXISTS_LOCAL=true
                        fi
                        
                        # Check remote tags
                        if git ls-remote --tags origin | grep -q "refs/tags/${VERSION_TAG}$"; then
                            TAG_EXISTS_REMOTE=true
                        fi
                        
                        # If tag exists anywhere, fail
                        if [ "$TAG_EXISTS_LOCAL" = true ] || [ "$TAG_EXISTS_REMOTE" = true ]; then
                            echo "Error: Tag ${VERSION_TAG} already exists"
                            if [ "$TAG_EXISTS_LOCAL" = true ]; then
                                echo "  - Tag exists locally"
                            fi
                            if [ "$TAG_EXISTS_REMOTE" = true ]; then
                                echo "  - Tag exists on remote (origin)"
                            fi
                            exit 1
                        else
                            echo "✓ Tag ${VERSION_TAG} does not exist yet (checked both local and remote)"
                        fi
                    '''
                }
            }
        }

        stage('Create Git Tag') {
            steps {
                withCredentials([string(credentialsId: 'github_token_for_terraform', variable: 'GITHUB_TOKEN')]) {
                    sh '''
                        git config user.email "jenkins@biot-med.com"
                        git config user.name "Jenkins CI"

                        git tag -a "${VERSION_TAG}" -m "Release ${VERSION_TAG}"

                        git push https://$GITHUB_TOKEN@github.com/biot-med/terraform-provider-biot-gen2.git ${VERSION_TAG}
                        echo "Pushed tag to remote"
                    '''
                }
            }
        }

        stage('Release with GoReleaser') {
            steps {
                script {
                    echo "Starting GoReleaser release for version ${env.VERSION_TAG}..."
                    
                    sh '''
                        # Ensure PATH includes common install locations
                        export PATH="$HOME/.local/bin:/usr/local/bin:$PATH"
                        
                        # Find goreleaser if not in PATH
                        if ! command -v goreleaser &> /dev/null; then
                            if [ -f "$HOME/.local/bin/goreleaser" ]; then
                                export PATH="$HOME/.local/bin:$PATH"
                            elif [ -f "/usr/local/bin/goreleaser" ]; then
                                export PATH="/usr/local/bin:$PATH"
                            else
                                echo "Error: goreleaser not found. Please ensure it is installed."
                                exit 1
                            fi
                        fi
                        
                        # Verify goreleaser is available
                        which goreleaser
                        goreleaser --version
                        
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

