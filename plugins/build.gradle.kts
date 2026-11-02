plugins {
    alias(libs.plugins.kotlin) apply false
    alias(libs.plugins.pluginPublish) apply false
}

allprojects {
    group = "com.sentiary"
    version = "2.1.0"
}
