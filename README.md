# Particle Life: Yeast-like Simulation

A high-performance particle interaction simulation written in Go. By defining simple attraction and repulsion rules between different colored groups, the simulation produces emergent life-like structures, flocking behaviors, and organic cellular patterns.

---

### 🖼️ Simulation Preview

![Yeast Simulation Screenshot]([https://github.com/arcesoftware/Particles/blob/main/Screenshot%202026-02-25%20134801.png])

---

## 🔬 How it Works

The simulation manages **6,600 particles** divided into three distinct groups:
* **Green (Type 0):** Representing yeast-like organic growth.
* **Red (Type 1):** Competitor particles.
* **Yellow (Type 2):** Partially compatible transition particles.

### The Physics of "Life"
Each particle follows a simple set of rules based on its neighbors within an **80-pixel interaction radius**. The force applied is determined by a gravity constant ($g$):

* **$g > 0$**: Particles repel each other.
* **$g < 0$**: Particles attract each other.

To keep the simulation stable and "liquid," a friction coefficient of `0.5` is applied to velocities every frame, simulating a viscous environment like a petri dish.



## 🚀 Features

* **Emergent Behavior:** Watch as random noise organizes into clusters, strings, and "cells."
* **Boundary Physics:** Particles feature elastic collisions with the window edges.
* **Customizable DNA:** Easily change the behavior of the "species" by modifying the interaction matrix.

## 🛠️ Installation & Running

1.  **Prerequisites:** * Go (1.18+)
    * SDL2 development libraries (required for the canvas driver)
2.  **Clone the repo:**
    ```bash
    git clone [https://github.com/your-username/plasma-yeast-sim.git](https://github.com/your-username/plasma-yeast-sim.git)
    cd plasma-yeast-sim
    ```
3.  **Run:**
    ```bash
    go run main.go
    ```

## ⚙️ Modifying Rules

You can change the "behavioral DNA" of the particles in the `main.go` file by adjusting the `rule` function parameters:

```go
// rule(source, target, gravity)
rule(0, 0, -0.05) // Green attracts Green
rule(0, 1, 0.10)  // Green repels Red
