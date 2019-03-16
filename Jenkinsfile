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
            def buildCache = "eu.gcr.io/${project}/${appName}-build-test:latest"
            sh """
            docker pull "${buildCache}" || "No docker cache found" || true
            docker build --cache-from ${buildCache} -t ${buildCache} -f Dockerfile-test .
            docker push ${buildCache}
            docker tag ${buildCache} testcontainer
            """
        }
    }

    stage('Identify LOC') {
        container('docker') {
            try {
                sh("""
                docker run -v \$(pwd):/data --rm -t xcnt/cloc-sloccount-wrapper:stable --not-match-f="(cloc.xml|swagger.*|cover.out|coverage.xml|xunit.xml)" --exclude-d vendor --xml --out=cloc.xml .
                """)
            } finally {
                sloccountPublish(pattern: 'sloccount.sc')
            }
        }
    }

    stage('Run Unit Tests') {
        container('docker') {
            try {
                sh("""
                docker run -v \$(pwd):/app --rm -t testcontainer make xunit
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
                docker run -v \$(pwd):/app --rm -t testcontainer make lint
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
        sed -i 's/Version = .*/Version = \"${versionPrefix}-\$(date +%Y%m%d%H%m%S)\"/' cli/version.go
        """

        container('docker') {
            imageWithTag = buildImage(image, env.BRANCH_NAME, myRepo.GIT_COMMIT, "build-env")
        }
    }

    stage('Publish') {
      publishImage(image, env.BRANCH_NAME, myRepo.GIT_COMMIT)
      if(env.BRANCH_NAME == "master") {
        publishImageToPublicDocker("xcnt/kubernetes-update-manager", image, env.BRANCH_NAME, myRepo.GIT_COMMIT)
      }
    }

    stage('Deploy') {
        loginToDocker()
        def updateClassifier = env.BRANCH_NAME
        if(updateClassifier == "master") {
            updateClassifier = "stable"
        }

        deployXCNTImage(imageWithTag, updateClassifier)
    }
}
