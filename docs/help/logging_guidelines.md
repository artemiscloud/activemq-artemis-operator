# Logging Guidlines

When you need add logging to your code, please consider the following rules:

* The operator configures the root logger (i.e. ctrl.Log) in main.go, so each new logger should be created from it in order to inherit the properties of it.

* Each logical object (like struct type) should use it's own logger with a proper name. Normally the process is use ctrl.Log.WithName(name) to create the logger and pass it to the struct's constructor. Logging in any of the struct's member functions should be using that logger. If need to you may create a logger using logger.WithValues() in a member function to further describe your logger.

* If a struct logically owns a struct (i.e. the creator and user of it), the struct may pass its logger to it when creating an instance, or create a new logger from it using Logger.WithName(name) so that the logger name has some hierarchy.

* For static methods that don't belong to any controllers, try to place those methods outside any controller packages (like utils, common, etc under pkgs dir). If logging is needed, create a logger inside the function body using ctrl.Log (operator's root logger) with a proper name (like util_jolokia). This mostly guarantees the logger inherit the properties of root logger when being called.

* Logging Level usage:

  * Logging levels are limited to error, level-0 and level-1 info. (i.e. Logger.Error(), Logger.Info() and Logger.V(1).Info())
  
  * Using Logger.Info() for messages describing major events like actions of creating, updating or describing an major operation, for example reconciling, resource changes, etc. The message body shouldn't be too large (i.e messages are concise and with key-value pairs to print simple type variables)
  
  * Using Logger.V(1).Info() for messages that give further details of each operation, or printing large body of objects like whole statefulsets, large configmaps, etc
  
  * In production deployment, set the log-level to error to avoid unnecessary details to log except error messages
  
  * In dev mode you can set log-level to info or debug to have details to show in logs in order to investigate an issue.
