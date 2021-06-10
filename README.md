## Conjure

Conjure up configuration values in your own defined template files.


### How does this work?

You create your configuration (.yml | json | toml) `template` to whichever structure
you want it to be in. Then you create the variables which you would like
to inject into the template file. This works perfectly for on the fly configuration
generation with secrets you receive over the network or from files that 
are encrypted/decrypted on runtime.

TL;DR

- You have a couple of configurations everyone can set. 
  
- You have configurations which only a select few can set - such as secrets.

- There are multiple files of the same structure with different values (production, staging etc.)

`Conjure` solves these use cases by allowing you to inherit files from one another
and inject values into your templates on your ci/cd.

Example

`conjure_parent.yml` is a `conjure` holding secrets

```yaml

files:
  - id: config
    path: templates/config.yml
    output: <tags>
  
tags:
  - id: production
    path: output/config.production.yml
  - id: development
    path: output/config.development.yml

groups:
  - id: conjure<production>
    items:
      - id: secret
        value: productionsecret
  
  - id: conjure<development>
    items:
      - id: secret
        value: developmentsecret
```

`conjure.yml` is a `conjure` holding accessible configs

```yaml
inherit: conjure_parent.yml # helps Conjure inject values from the parent

files:
  - id: config
    path: templates/config.yml
    output: <tags>
  - id: config
    port: templates/config.yml
    output: output/config.default.yml

tags:
  - id: production
    path: output/config.production.yml
  - id: development
    path: output/config.development.yml

groups:
  - id: conjure<production>
    items:
      - id: host
        value: 0.0.0.0
      - id: port
        value: 8753
  
  - id: conjure<development>
    items:
      - id: host
        value: 127.0.0.1
      - id: port
        value: 8080

  - id: conjure
    items:
      - id: host
        value: 193.241.344.111
      - id: port
        value: 64356
```

`conjure-alt.yml` is a `conjure` holding alternative configs
```yaml
inherit: conjure.yml

files:
  - id: config-alt
    path: templates/config.yml # using the same template
    output: <tags> # also using the same tags
```

`config.yml` is your `template` structure
```yaml
host: ${conjure.host|conjure-alt.host}
port: ${conjure.port}
secret: ${conjure.secret}
```

#### Use case 1

You have a .git repository which has a ton of configurations needed for 
different use cases such as a configuration file for your production environment,
another for your development environment and maybe even one for staging, or a second
production environment.

The problem here is that not only do you have these files in the same structure (usually)
you have different values for each file - hopefully ;).
This can become problematic as there isn't a single source of truth from which
you can easily compare values or easily change them. 

Another problem is that each of these files might have different secret
configurations which only certain people have access to. This makes it tough
to have others configure files with secrets they don't even know about.

This is where `Conjure` can help with creating `conjures` that hold your
secrets or/and configurations. This ensures that others can continue working
on setting configuration values without touching secrets or configs only a select
few are allowed to set.


### Conjure syntax

Conjure uses a `${}` to indicate `template` placeholders.
The values inside the placeholder is formatted according to
your `group` and then the `item`.

e.g. `${group-id.item_id}`

Inside your `conjure` files you can set up defaults and tags.

Tags are used to populate more than one file without the need to set up
more variables.

```yaml
files:
  - id: config
    path: templates/config.yml
    output: <tags> # this will use all the tags listed under `tags` and populate a new file under each of their 
                   # default directories
    
tags:
  - id: production # your custom tag
    path: output/config.production.yml # the default output directory for this tag
  - id: development # another custom tag
    path: output/config.development.yml

groups:
  - id: conjure<production> # create a tag specific variables
```

#### Inheritance

In the case of inheritance the files listed inside the parent `conjure`
are ignored.

```yaml
inherit: conjure_parent.yml
```

`conjure_parent.yml`

```yaml
# these here will be ignored by the child
# call this file directly for its files to be rendered.

files:
  - id: config
    path: templates/config.yml
    output: <tags>
```