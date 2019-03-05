@Library('xcnt-jenkins-scripts') _

def project = 'xcnt-infrastructure'
def appName = 'kubernetes-update-manager'
def label = "${appName}_${env.BRANCH_NAME}"

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
                sh("""
                docker run -v \$(pwd):/app --rm -t testcontainer bash scripts/run-xunit-tests.sh
                """)
            }
            finally {
                junit '**/xunit.xml'
                cobertura coberturaReportFile: '**/coverage.xml'
            }
        }
    }

    stage('Run Go Checks') {
        try {
            container('docker') {
                sh """
                docker run -v \$(pwd):/app --rm -t testcontainer bash scripts/run-golint.sh
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
