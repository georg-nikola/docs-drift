# Example Documentation

This is an example markdown file with code blocks.

## JavaScript Example

```javascript
const greeting = "Hello, World!";
console.log(greeting);
```

## Python Example

```python
greeting = "Hello, World!"
print(greeting)
```

## Skipped Example

```javascript docs-drift:skip
// This should not be executed
throw new Error("Should be skipped");
```

## Another Valid Example

```javascript
// This works correctly
const numbers = [1, 2, 3];
const doubled = numbers.map(n => n * 2);
console.log(doubled);
```
