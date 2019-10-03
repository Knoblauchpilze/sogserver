
# include <core_utils/StdLogger.hh>
# include <core_utils/LoggerLocator.hh>

# include <core_utils/CoreException.hh>

int main(int argc, char* argv[]) {
  // Create the logger.
  utils::StdLogger logger;
  utils::LoggerLocator::provide(&logger);

  const std::string service("mandelbulb");
  const std::string module("main");

  try {
    // Implement something in here.
  }
  catch (const utils::CoreException& e) {
    utils::LoggerLocator::getLogger().logMessage(
      utils::Level::Critical,
      std::string("Caught internal exception while running sog server"),
      module,
      service,
      e.what()
    );
  }
  catch (const std::exception& e) {
    utils::LoggerLocator::getLogger().logMessage(
      utils::Level::Critical,
      std::string("Caught exception while running sog server"),
      module,
      service,
      e.what()
    );
  }
  catch (...) {
    utils::LoggerLocator::getLogger().logMessage(
      utils::Level::Critical,
      std::string("Unexpected error while running sog server"),
      module,
      service
    );
  }

  // All is good.
  return EXIT_SUCCESS;
}
