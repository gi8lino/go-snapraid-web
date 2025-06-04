function attachFilter({ inputId, rowSelector, getText }) {
  const input = document.getElementById(inputId);
  if (!input) return;

  const wrapper = input.closest(".input-group.filter");
  const clearBtn = wrapper?.querySelector(".clear-filter-btn");

  const filterRows = () => {
    const needle = input.value.trim().toLowerCase();
    document.querySelectorAll(rowSelector).forEach((row) => {
      const txt = getText(row).trim().toLowerCase();
      row.style.display = !needle || txt.includes(needle) ? "" : "none";
    });
  };

  input.addEventListener("input", () => {
    if (clearBtn) clearBtn.style.display = input.value ? "block" : "none";
    filterRows();
  });

  if (clearBtn) {
    clearBtn.addEventListener("click", () => {
      input.value = "";
      clearBtn.style.display = "none";
      input.focus();
      document.querySelectorAll(rowSelector).forEach((row) => {
        row.style.display = "";
      });
    });

    clearBtn.style.display = input.value ? "block" : "none";
  }
}

async function loadSection(sec) {
  try {
    let url = `/partials/${sec}`;

    if (sec === "details") {
      const hash = window.location.hash.slice(1);
      const parts = hash.split("/");
      if (parts.length >= 3 && parts[1] === "details") {
        const rawId = decodeURIComponent(parts.slice(2).join("/"));
        url += `?id=${encodeURIComponent(rawId)}`;
      }
    }

    const res = await fetch(url);
    if (!res.ok) {
      document.getElementById("content").innerHTML =
        `<p class='text-danger'>Error ${res.status} loading ${sec}.</p>`;
      return;
    }

    const html = await res.text();
    document.getElementById("content").innerHTML = html;

    // Highlight active nav link
    const sectionName = sec.toLowerCase();
    document.querySelectorAll("nav .nav-link").forEach((a) => {
      a.classList.toggle(
        "active",
        a.dataset.section?.toLowerCase() === sectionName,
      );
    });

    if (sec === "overview") {
      document.querySelectorAll("#overview tbody tr").forEach((row) => {
        const dateCell = row.querySelector("td[data-timestamp]");
        if (dateCell) {
          dateCell.style.cursor = "pointer";
          dateCell.addEventListener("click", () => {
            const ts = dateCell.getAttribute("data-timestamp");
            goToDetails(ts);
          });
        }
      });

      attachFilter({
        inputId: "searchOverview",
        rowSelector: "#overview tbody tr",
        getText: (row) => row.querySelector("td:nth-child(1)").textContent,
      });

      const table = document.querySelector("#overview table");
      if (table && window.Tablesort) new Tablesort(table);
    }

    if (sec === "details") {
      const selector = document.getElementById("runSelector");
      if (selector) {
        selector.addEventListener("change", (e) => {
          const newId = e.target.value;
          goToDetails(newId);
        });
      }
    }
  } catch (err) {
    console.error("loadSection error:", err);
    document.getElementById("content").innerHTML =
      "<p class='text-danger'>Unexpected error.</p>";
  }
}

async function goToDetails(id) {
  window.location.hash = `/details/${encodeURIComponent(id)}`;
  await loadSection("details");
}

document.addEventListener("DOMContentLoaded", () => {
  document.querySelectorAll("nav .nav-link").forEach((a) => {
    a.addEventListener("click", (e) => {
      e.preventDefault();
      const sec = a.dataset.section.toLowerCase();
      if (sec) {
        window.location.hash = `#/${sec}`;
        loadSection(sec);
      }
    });
  });

  const initial = window.location.hash.slice(1);
  if (!initial || initial === "/overview") {
    window.location.hash = "/overview";
    loadSection("overview");
  } else if (initial.startsWith("/details")) {
    loadSection("details");
  } else {
    window.location.hash = "/overview";
    loadSection("overview");
  }
});
