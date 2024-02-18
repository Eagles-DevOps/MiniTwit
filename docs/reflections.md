# Reflections

## Reasons for choosing Go

- Lower memory consumption compared to ruby, Java etc.
- Scales well.
    - which is important for us considering the userbase will increase over time.
- Easy to learn.
- Compiles to a single binary which makes it easy to deploy (compared to).
- Comes with an opinionated formatter (gofmt) ensuring code style consistency. This reduces discussions about styling (which does not bring value to the product).
- It is a language with high developer interest meaning that it will hopefully be ease future recruitment when the application sky-rockets in popularity.

## Server management considerations
Some open questions regarding how we want we intend to manage the server:
- Should we host the applicaiton in docker on the server?
- Should we work from the principle of immutable infrastructure?
- Where should we host it?
- Server vs managed container service?
