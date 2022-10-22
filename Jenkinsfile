def vers
def outFile
def release = false
pipeline {
    agent any
    tools {
        go 'Go 1.19'
        maven 'Mvn'
    }
    environment {
        NEXUS_CREDS = credentials('Cantara-NEXUS')
    }
    stages {
        stage("pre") {
            steps {
                script {
                    if (env.TAG_NAME) {
                        vers = "${env.TAG_NAME}"
                        release = true
                    } else {
                        vers = "${env.GIT_COMMIT}"
                    }
                    outFile = "wamper-${vers}"
                    echo "New file: ${outFile}"
                }
            }
        }
        stage("test") {
            steps {
                script {
                    testApp()
                }
            }
        }
        stage("build") {
            steps {
                script {
                    echo "V: ${vers}"
                    echo "File: ${outFile}"
                    buildApp(outFile, vers)
                }
            }
        }
        stage("deploy") {
            steps {
                script {
                    echo 'deplying the application...'
                    echo "deploying version ${vers}"
                    if (release) {
                        sh 'curl -v -u $NEXUS_CREDS '+"--upload-file ${outFile} https://mvnrepo.cantara.no/content/repositories/releases/no/cantara/gotools/wamper/${vers}/${outFile}"
                    } else {
                        sh 'curl -v -u $NEXUS_CREDS '+"--upload-file ${outFile} https://mvnrepo.cantara.no/content/repositories/snapshots/no/cantara/gotools/wamper/${vers}/${outFile}"
                    }
                    sh "rm ${outFile}"
                }
            }
        }
    }
}

def testApp() {
    echo 'testing the application...'
    sh './testRecursive.sh'
}

def buildApp(outFile, vers) {
    echo 'building the application...'
    sh 'ls'
    sh "CGO_ENABLED=0 GOOD=linux GOARCH=amd64 go build -ldflags \"-X 'github.com/cantara/wamper/health.Version=${vers}' -X 'github.com/cantara/wamper/health.BuildTime=\$(date)' -X 'github.com/cantara/wamper/web.Name=wamper'\" -o ${outFile}"
}
