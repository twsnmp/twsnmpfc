<template>
  <img class="Logo" alt="TWSNMP FC Logo" src="/images/twsnmp.png" />
</template>

<style>
.Logo {
  height: 200px;
  width: 200px;
  background-color: ghostwhite;
  transform: rotateY(560deg);
  animation: turn 3.5s ease-out forwards 1s;
}

@keyframes turn {
  100% {
    transform: rotateY(0deg);
  }
}
</style>
