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
rootProject.name = "sentiary-mobile-example"
include(":android:app", ":shared")

sentiary {
    android {
        create("app") {
            modulePath = ":android:app"
            output = "android"
            sourceSet = "main"
            resourceRoot = "android/app/build/generated/sentiary"
        }
    }
    compose {
        create("shared") {
            modulePath = ":shared"
            output = "compose"
            sourceSet = "commonMain"
            resourceRoot = "shared/build/generated/sentiary/composeResources"
        }
    }
}
