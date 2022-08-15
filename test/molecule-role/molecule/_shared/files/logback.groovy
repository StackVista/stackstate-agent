// This can be used for debugging
// statusListener(OnConsoleStatusListener)

try {
    // Use the stupid scanner trick to read a resource as string without any dependencies
    String fileContent = new Scanner(getClass().getResourceAsStream('/logback.base.groovy'), "UTF-8").useDelimiter("\\A").next();
    // Poor man's include, we read the file and instantiate it with the gaffer groovy executor
    // we do this to keep the default configuration and make this file extensible in kubernetes
    new ch.qos.logback.classic.gaffer.GafferConfigurator(getContext()).run(fileContent)
} catch (Exception e) {
    println "Error including groovy base file " + e.toString()
}

def defaultLogPattern = "%date %-5level %logger{60} - %msg%n"

appender("Console", ConsoleAppender) {
    encoder(PatternLayoutEncoder) {
        pattern = defaultLogPattern
    }
}

root(DEBUG, ["Console"])
