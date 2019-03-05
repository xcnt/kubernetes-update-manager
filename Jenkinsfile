@Library('xcnt-jenkins-scripts') _

def project = 'xcnt-infrastructure'
def appName = 'kubernetes-update-manager'
def label = "${appName}_${env.BRANCH_NAME}"

@NonCPS
Boolean includesTest(File directory) {
    return new File("${directory.GetName()}/common_test.go").exists()
}

@NonCPS
Set<String> findTestDirs(String dirName)
{
    Set dirList = []
    new File(workspace).eachDir()
    {
        dir -> 
        if (!dir.getName().trim.endsWith('vendor')) {
            if (!dir.getName().startsWith('.') && includesTest(dir)) {
                dirList.add(dir.getName())
            }
            dirList.addAll(findTestDirs(dir.getName()))
        }
    }
    dirList
}

dockerBuildRuntime(label: label) {
    def myRepo = checkout scm
    def image = "eu.gcr.io/${project}/${appName}"
    def imageWithTag = ""

    loginToDocker()

    stage('Build test image') {
        container('docker') {
            sh """
            docker build -t testcontainer -f Dockerfile-test .
            """
        }
    }

    stage('Run Unit Tests') {
        container('docker') {
            try {
                def directories = findTestDirs('.')
                directories.each { testDirectory -> 
                    xunit = sh(returnStdout: true, script: """
                                                docker run -t testcontainer -w "${testDirectory}" "go test -gocheck.vv | go2xunit -gocheck"
                                                """).trim()
                    writeFile(file: "${testDirectory}/xunit.xml", text: xunit)
                }
            }
            finally {
                junit '**/xunit.xml'
            }
        }
    }

    stage('Run Go Checks') {
        try {
            container('docker') {
                sh """
                docker run -t testcontainer go vet ./... > govet.xml
                docker run -t testcontainer golint ./... > golint.xml
                """
            }
        } finally {
            recordIssues enabledForFailure: true, tool: goVet(pattern: '**/govet.xml'), qualityGates: [[threshold: 1, type: 'TOTAL', unstable: true]]
            recordIssues enabledForFailure: true, tool: goLint(pattern: '**/golint.xml'), qualityGates: [[threshold: 1, type: 'TOTAL', unstable: true]]
        }
    }

    stage('Build image') {
        container('docker') {
            imageWithTag = buildImage(image, env.BRANCH_NAME, myRepo.GIT_COMMIT)
        }
    }

    stage('Publish') {
      publishImage(image, env.BRANCH_NAME, myRepo.GIT_COMMIT)
    }
}
