## Kubewizard (working title)

A POC to investigate generating controller logic from a 
visual representation.

## Motivation

Operators automate the processes of managing applications
usually carried out or scripted by subject matter experts
in a given domain. A common problem when developing
operators is these subject matter experts not having the
skillset to develop the operators and the people that
do have the skillset rarely have the domain knowledge 
required to develop the operator which can slow down 
development time and increase the risk of mistakes.

## Objective

Enable subject matter experts and operator developers to
communicate and develop high level controller logic visually.

## User Stories

### Subject Matter Expert

- As a subject matter expert I can compose a process visually
- As a subject matter expert I can recompose a process visually

### Developer

- As a developer I can use processes composed by subject matter
experts to scaffold code
- As a developer I can use a process composed by a subject matter
expert to rescaffold code

## Constraints

- Processes should be easily composed and recomposed
- Recomposing a process should have minimal impact on developers

## Implementation

Kubewizard uses [kubebuilder] to scaffold controllers and
resources and also implements the logic defined in the [bpmn]
diagram specified when scaffolding the controller. This allows
a subject matter expert to compose and recompose processes
easily and for it to be implemented in go for further development.

[kubebuilder]:https://github.com/kubernetes-sigs/kubebuilder
[bpmn]:http://www.bpmn.org
