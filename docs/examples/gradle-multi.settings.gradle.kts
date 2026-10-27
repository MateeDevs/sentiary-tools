// Unreleased Gradle plugin 2.0: replace this path with the plugin source checkout.
pluginManagement {
    includeBuild("PATH_TO_PLUGIN_SOURCE")
    repositories {
        google()
        mavenCentral()
        gradlePluginPortal()
    }
}
plugins {
    id("com.sentiary.gradle")
}
rootProject.name = "sentiary-multi-example"
include(":app", ":library")

sentiary {
    android {
        create("appMain") {
            modulePath = ":app"
            output = "app_main"
            sourceSet = "main"
            resourceRoot = "app/build/generated/sentiary/main"
        }
        create("appDebug") {
            modulePath = ":app"
            output = "app_debug"
            sourceSet = "debug"
            resourceRoot = "app/build/generated/sentiary/debug"
        }
        create("libraryMain") {
            modulePath = ":library"
            output = "library_main"
            sourceSet = "main"
            resourceRoot = "library/build/generated/sentiary/main"
        }
    }
}
