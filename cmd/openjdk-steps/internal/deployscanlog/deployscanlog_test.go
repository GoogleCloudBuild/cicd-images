// Copyright 2023 Google LLC All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package deployscanlog

import (
	"reflect"
	"testing"
)

func TestRetrieveArtifacts(t *testing.T) {
	// Define test cases
	log :=  "mvn deploy [INFO] Scanning for projects... [WARNING] [WARNING] Some problems were encountered while building the effective model for com.mycompany.app:cloudbuild-test-maven:jar:1.0-SNAPSHOT [WARNING] 'build.plugins.plugin.version' for org.cyclonedx:cyclonedx-maven-plugin is missing. @ line 92, column 15 [WARNING] [WARNING] It is highly recommended to fix these problems because they threaten the stability of your build. [WARNING] [WARNING] For this reason, future Maven versions might no longer support building such malformed projects. [WARNING] [INFO] [INFO] --------------< com.mycompany.app:cloudbuild-test-maven >--------------- [INFO] Building cloudbuild-test-maven 1.0-SNAPSHOT [INFO] from pom.xml [INFO] --------------------------------[ jar ]--------------------------------- [INFO] [INFO] --- enforcer:3.2.1:enforce (enforce-maven) @ cloudbuild-test-maven --- [INFO] Rule 0: org.apache.maven.enforcer.rules.version.RequireMavenVersion passed [INFO] Rule 1: org.apache.maven.enforcer.rules.version.RequireJavaVersion passed [INFO] [INFO] --- resources:3.3.1:resources (default-resources) @ cloudbuild-test-maven --- [INFO] skip non existing resourceDirectory /Users/yawenluo/Desktop/testfolder/local_dev/simple-java-app/src/main/resources [INFO] [INFO] --- compiler:3.11.0:compile (default-compile) @ cloudbuild-test-maven --- [INFO] Nothing to compile - all classes are up to date [INFO] [INFO] --- resources:3.3.1:testResources (default-testResources) @ cloudbuild-test-maven --- [INFO] skip non existing resourceDirectory /Users/yawenluo/Desktop/testfolder/local_dev/simple-java-app/src/test/resources [INFO] [INFO] --- compiler:3.11.0:testCompile (default-testCompile) @ cloudbuild-test-maven --- [INFO] Nothing to compile - all classes are up to date [INFO] [INFO] --- surefire:3.1.2:test (default-test) @ cloudbuild-test-maven --- [INFO] Using auto detected provider org.apache.maven.surefire.junitplatform.JUnitPlatformProvider [INFO] [INFO] ------------------------------------------------------- [INFO] T E S T S [INFO] ------------------------------------------------------- [INFO] Running com.mycompany.app.AppTest [INFO] Tests run: 2, Failures: 0, Errors: 0, Skipped: 0, Time elapsed: 0.060 s -- in com.mycompany.app.AppTest [INFO] [INFO] Results: [INFO] [INFO] Tests run: 2, Failures: 0, Errors: 0, Skipped: 0 [INFO] [INFO] [INFO] --- jar:3.3.0:jar (default-jar) @ cloudbuild-test-maven --- [INFO] [INFO] --- cyclonedx:2.7.9:makeAggregateBom (default) @ cloudbuild-test-maven --- [INFO] CycloneDX: Resolving Dependencies [INFO] CycloneDX: Creating BOM version 1.4 with 0 component(s) [INFO] CycloneDX: Writing and validating BOM (XML): /Users/yawenluo/Desktop/testfolder/local_dev/simple-java-app/target/bom.xml [INFO] attaching as cloudbuild-test-maven-1.0-SNAPSHOT-cyclonedx.xml [INFO] CycloneDX: Writing and validating BOM (JSON): /Users/yawenluo/Desktop/testfolder/local_dev/simple-java-app/target/bom.json [WARNING] Unknown keyword additionalItems - you should define your own Meta Schema. If the keyword is irrelevant for validation, just use a NonValidationKeyword [INFO] attaching as cloudbuild-test-maven-1.0-SNAPSHOT-cyclonedx.json [INFO] [INFO] --- install:3.1.1:install (default-install) @ cloudbuild-test-maven --- [INFO] Installing /Users/yawenluo/Desktop/testfolder/local_dev/simple-java-app/pom.xml to /Users/yawenluo/.m2/repository/com/mycompany/app/cloudbuild-test-maven/1.0-SNAPSHOT/cloudbuild-test-maven-1.0-SNAPSHOT.pom [INFO] Installing /Users/yawenluo/Desktop/testfolder/local_dev/simple-java-app/target/cloudbuild-test-maven-1.0-SNAPSHOT.jar to /Users/yawenluo/.m2/repository/com/mycompany/app/cloudbuild-test-maven/1.0-SNAPSHOT/cloudbuild-test-maven-1.0-SNAPSHOT.jar [INFO] Installing /Users/yawenluo/Desktop/testfolder/local_dev/simple-java-app/target/bom.xml to /Users/yawenluo/.m2/repository/com/mycompany/app/cloudbuild-test-maven/1.0-SNAPSHOT/cloudbuild-test-maven-1.0-SNAPSHOT-cyclonedx.xml [INFO] Installing /Users/yawenluo/Desktop/testfolder/local_dev/simple-java-app/target/bom.json to /Users/yawenluo/.m2/repository/com/mycompany/app/cloudbuild-test-maven/1.0-SNAPSHOT/cloudbuild-test-maven-1.0-SNAPSHOT-cyclonedx.json [INFO] [INFO] --- deploy:3.1.1:deploy (default-deploy) @ cloudbuild-test-maven --- [INFO] Downloading from artifact-registry: https://us-central1-maven.pkg.dev/gcb-catalog-experiment/quickstart-java-repo/com/mycompany/app/cloudbuild-test-maven/1.0-SNAPSHOT/maven-metadata.xml [INFO] Downloaded from artifact-registry: https://us-central1-maven.pkg.dev/gcb-catalog-experiment/quickstart-java-repo/com/mycompany/app/cloudbuild-test-maven/1.0-SNAPSHOT/maven-metadata.xml (1.2 kB at 854 B/s) [INFO] Uploading to artifact-registry: https://us-central1-maven.pkg.dev/gcb-catalog-experiment/quickstart-java-repo/com/mycompany/app/cloudbuild-test-maven/1.0-SNAPSHOT/cloudbuild-test-maven-1.0-20230919.213015-4.pom [INFO] Uploaded to artifact-registry: https://us-central1-maven.pkg.dev/gcb-catalog-experiment/quickstart-java-repo/com/mycompany/app/cloudbuild-test-maven/1.0-SNAPSHOT/cloudbuild-test-maven-1.0-20230919.213015-4.pom (3.4 kB at 4.0 kB/s) [INFO] Uploading to artifact-registry: https://us-central1-maven.pkg.dev/gcb-catalog-experiment/quickstart-java-repo/com/mycompany/app/cloudbuild-test-maven/1.0-SNAPSHOT/cloudbuild-test-maven-1.0-20230919.213015-4.jar [INFO] Uploading to artifact-registry: https://us-central1-maven.pkg.dev/gcb-catalog-experiment/quickstart-java-repo/com/mycompany/app/cloudbuild-test-maven/1.0-SNAPSHOT/cloudbuild-test-maven-1.0-20230919.213015-4-cyclonedx.xml [INFO] Uploading to artifact-registry: https://us-central1-maven.pkg.dev/gcb-catalog-experiment/quickstart-java-repo/com/mycompany/app/cloudbuild-test-maven/1.0-SNAPSHOT/cloudbuild-test-maven-1.0-20230919.213015-4-cyclonedx.json [INFO] Uploading to artifact-registry: https://us-central1-maven.pkg.dev/gcb-catalog-experiment/quickstart-java-repo/com/mycompany/app/cloudbuild-test-maven/1.0-SNAPSHOT/CDE.war (3.6 kB at 4.0 kB/s) [INFO] Uploaded to artifact-registry: https://us-central1-maven.pkg.dev/gcb-catalog-experiment/quickstart-java-repo/com/mycompany/app/cloudbuild-test-maven/1.0-SNAPSHOT/cloudbuild-test-maven-1.0-20230919.213015-4.jar (3.6 kB at 4.0 kB/s) [INFO] Uploaded to artifact-registry: https://us-central1-maven.pkg.dev/gcb-catalog-experiment/quickstart-java-repo/com/mycompany/app/cloudbuild-test-maven/1.0-SNAPSHOT/cloudbuild-test-maven-1.0-sources.jar (3.6 kB at 4.0 kB/s) [INFO] Uploaded to artifact-registry: https://us-central1-maven.pkg.dev/gcb-catalog-experiment/quickstart-java-repo/com/mycompany/app/cloudbuild-test-maven/1.0-SNAPSHOT/ABC.jar (3.6 kB at 4.0 kB/s) [INFO] Uploaded to artifact-registry: https://us-central1-maven.pkg.dev/gcb-catalog-experiment/quickstart-java-repo/com/mycompany/app/cloudbuild-test-maven/1.0-SNAPSHOT/CDE.war (3.6 kB at 4.0 kB/s) [INFO] Uploaded to artifact-registry: https://us-central1-maven.pkg.dev/gcb-catalog-experiment/quickstart-java-repo/com/mycompany/app/cloudbuild-test-maven/1.0-SNAPSHOT/CDE.ear (3.6 kB at 4.0 kB/s) [INFO] Uploaded to artifact-registry: https://us-central1-maven.pkg.dev/gcb-catalog-experiment/quickstart-java-repo/com/mycompany/app/cloudbuild-test-maven/1.0-SNAPSHOT/CDE.rar (3.6 kB at 4.0 kB/s) [INFO] Uploaded to artifact-registry: https://us-central1-maven.pkg.dev/gcb-catalog-experiment/quickstart-java-repo/com/mycompany/app/cloudbuild-test-maven/1.0-SNAPSHOT/cloudbuild-test-maven-1.0-20230919.213015-4-cyclonedx.xml (2.3 kB at 2.4 kB/s) [INFO] Uploaded to artifact-registry: https://us-central1-maven.pkg.dev/gcb-catalog-experiment/quickstart-java-repo/com/mycompany/app/cloudbuild-test-maven/1.0-SNAPSHOT/cloudbuild-test-maven-1.0-20230919.213015-4-cyclonedx.json (2.7 kB at 2.7 kB/s) [INFO] Downloading from artifact-registry: https://us-central1-maven.pkg.dev/gcb-catalog-experiment/quickstart-java-repo/com/mycompany/app/cloudbuild-test-maven/maven-metadata.xml [INFO] Downloaded from artifact-registry: https://us-central1-maven.pkg.dev/gcb-catalog-experiment/quickstart-java-repo/com/mycompany/app/cloudbuild-test-maven/maven-metadata.xml (352 B at 1.1 kB/s) [INFO] Uploading to artifact-registry: https://us-central1-maven.pkg.dev/gcb-catalog-experiment/quickstart-java-repo/com/mycompany/app/cloudbuild-test-maven/1.0-SNAPSHOT/maven-metadata.xml [INFO] Uploaded to artifact-registry: https://us-central1-maven.pkg.dev/gcb-catalog-experiment/quickstart-java-repo/com/mycompany/app/cloudbuild-test-maven/1.0-SNAPSHOT"
	expectedResult := map[string]string{
				"https://us-central1-maven.pkg.dev/gcb-catalog-experiment/quickstart-java-repo/com/mycompany/app/cloudbuild-test-maven/1.0-SNAPSHOT/ABC.jar":                                         "https://us-central1-maven.pkg.dev/gcb-catalog-experiment/quickstart-java-repo/com/mycompany/app/cloudbuild-test-maven/1.0-SNAPSHOT/ABC.jar",
				"https://us-central1-maven.pkg.dev/gcb-catalog-experiment/quickstart-java-repo/com/mycompany/app/cloudbuild-test-maven/1.0-SNAPSHOT/cloudbuild-test-maven-1.0-sources.jar":           "https://us-central1-maven.pkg.dev/gcb-catalog-experiment/quickstart-java-repo/com/mycompany/app/cloudbuild-test-maven/1.0-SNAPSHOT/cloudbuild-test-maven-1.0-sources.jar",
				"https://us-central1-maven.pkg.dev/gcb-catalog-experiment/quickstart-java-repo/com/mycompany/app/cloudbuild-test-maven/1.0-SNAPSHOT/CDE.war":                                         "https://us-central1-maven.pkg.dev/gcb-catalog-experiment/quickstart-java-repo/com/mycompany/app/cloudbuild-test-maven/1.0-SNAPSHOT/CDE.war",
				"https://us-central1-maven.pkg.dev/gcb-catalog-experiment/quickstart-java-repo/com/mycompany/app/cloudbuild-test-maven/1.0-SNAPSHOT/CDE.ear":                                         "https://us-central1-maven.pkg.dev/gcb-catalog-experiment/quickstart-java-repo/com/mycompany/app/cloudbuild-test-maven/1.0-SNAPSHOT/CDE.ear",
				"https://us-central1-maven.pkg.dev/gcb-catalog-experiment/quickstart-java-repo/com/mycompany/app/cloudbuild-test-maven/1.0-SNAPSHOT/CDE.rar":                                         "https://us-central1-maven.pkg.dev/gcb-catalog-experiment/quickstart-java-repo/com/mycompany/app/cloudbuild-test-maven/1.0-SNAPSHOT/CDE.rar",
				"https://us-central1-maven.pkg.dev/gcb-catalog-experiment/quickstart-java-repo/com/mycompany/app/cloudbuild-test-maven/1.0-SNAPSHOT/cloudbuild-test-maven-1.0-20230919.213015-4.jar": "https://us-central1-maven.pkg.dev/gcb-catalog-experiment/quickstart-java-repo/com/mycompany/app/cloudbuild-test-maven/1.0-SNAPSHOT/cloudbuild-test-maven-1.0-20230919.213015-4.jar",
			}

	// Run test cases
	t.Run("Full log of mvn deploy", func(t *testing.T) {
		// Call function
		result := retrieveArtifacts(log)
		if !reflect.DeepEqual(result, expectedResult) {
			t.Errorf("Expected %v\n, but got %v", expectedResult, result)
		}
	})

}