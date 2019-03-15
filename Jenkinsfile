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

    stage('Identify LOC') {
        container('docker') {
            try {
                sh("""
                docker run -v \$(pwd):/data --rm -t mribeiro/cloc --not-match-f cloc.xml --exclude-d vendor --xml --out=cloc.xml .
                """)
            } finally {
                sloccountPublish(pattern: 'cloc.xml')
            }
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
        def versionPrefix = "dev"
        if (env.BRANCH_NAME == "master") {
            versionPrefix = "stable"
        }
        sh """
        sed -i 's/^Version = .*\$/Version = \"${versionPrefix}-\$(date +%Y%m%d%H%m%S)\"/' cli/version.go
        """

        container('docker') {
            imageWithTag = buildImage(image, env.BRANCH_NAME, myRepo.GIT_COMMIT)
        }
    }

    stage('Publish') {
      publishImage(image, env.BRANCH_NAME, myRepo.GIT_COMMIT)
    }
}
