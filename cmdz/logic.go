package cmdz

/*
# conditions
- Success(Executer)
- RcEquals(Executer, int)
- StdoutContains(Executer, string)
return a condition

# logic
- Not(condition)
- And(condition1, condition2, ...) execute all in sequence while condition is evaluated true
- Or(condition1, condition2, ...) execute all in sequence while condition is evaluated false
return an Executer

*/
