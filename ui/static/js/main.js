// Add live class to the current page link in the nav
const navLinks = document.querySelectorAll("nav a");
for (var i = 0; i < navLinks.length; i++) {
	const link = navLinks[i]
	if (link.getAttribute('href') == window.location.pathname) {
		link.classList.add("live");
		break;
	}
}

// Mobile navigation toggle
document.addEventListener("DOMContentLoaded", function () {
    const mobileNavToggle = document.querySelector("#mobile-nav-toggle");
    const mobileNavMenu = document.querySelector("#mobile-nav-menu");
    const hamburgerIcon = document.querySelector("#hamburger-icon");
    const closeIcon = document.querySelector("#close-icon");
    const body = document.querySelector("body");
    const mobileNavMenuOverlay = document.querySelector(
        "#mobile-nav-menu-overlay"
    );

    if (mobileNavToggle) {
        mobileNavToggle.addEventListener("change", function () {
            if (this.checked) {
                if (mobileNavMenu)
                    mobileNavMenu.classList.remove("invisible", "left-[100%]");
                if (mobileNavMenu)
                    mobileNavMenu.classList.add("visible", "left-0");
                if (hamburgerIcon) hamburgerIcon.classList.add("hidden");
                if (hamburgerIcon) hamburgerIcon.classList.remove("flex");
                if (closeIcon) closeIcon.classList.remove("hidden");
                if (closeIcon) closeIcon.classList.add("flex");
                if (body) body.classList.add("fixed");
                if (mobileNavMenuOverlay)
                    mobileNavMenuOverlay.classList.remove("hidden");
                if (mobileNavMenuOverlay)
                    mobileNavMenuOverlay.classList.add("flex");
            } else {
                if (mobileNavMenu)
                    mobileNavMenu.classList.add("invisible", "left-[100%]");
                if (mobileNavMenu)
                    mobileNavMenu.classList.remove("visible", "left-0");
                if (hamburgerIcon) hamburgerIcon.classList.remove("hidden");
                if (hamburgerIcon) hamburgerIcon.classList.add("flex");
                if (closeIcon) closeIcon.classList.add("hidden");
                if (closeIcon) closeIcon.classList.remove("flex");
                if (body) body.classList.remove("fixed");
                if (mobileNavMenuOverlay)
                    mobileNavMenuOverlay.classList.remove("flex");
                if (mobileNavMenuOverlay)
                    mobileNavMenuOverlay.classList.add("hidden");
            }
        });
    } else {
        console.warn("Mobile navigation toggle element not found");
    }
});

// Multiselect dropdown
document.addEventListener("DOMContentLoaded", function () {
    const multiselects = document.querySelectorAll(".multiselect-dropdown");

    multiselects.forEach((multiselect) => {
        const select = multiselect.previousElementSibling;
        const selectedOptions = multiselect.querySelector(".selected-options");
        const searchInput = multiselect.querySelector(".search-input");
        const optionsList = multiselect.querySelector(".options-list");

        // Populate options list
        select.querySelectorAll("option").forEach((option) => {
            const li = document.createElement("li");
            li.textContent = option.textContent;
            li.dataset.value = option.value;
            li.className = "p-2 hover:bg-gray-100 cursor-pointer";
            optionsList.appendChild(li);
        });

        // Toggle dropdown visibility
        searchInput.addEventListener("focus", () => {
            optionsList.style.display = "block";
        });

        // Close dropdown when clicking outside
        document.addEventListener("click", (e) => {
            if (!multiselect.contains(e.target)) {
                optionsList.style.display = "none";
            }
        });

        // Toggle option selection
        optionsList.addEventListener("click", (e) => {
            if (e.target.tagName === "LI") {
                e.target.classList.toggle("selected");
                updateSelectedOptions();
                e.stopPropagation(); // Prevent closing dropdown
            }
        });

        // Filter options based on search input
        searchInput.addEventListener("input", (e) => {
            const searchTerm = e.target.value.toLowerCase();
            optionsList.querySelectorAll("li").forEach((li) => {
                const optionText = li.textContent.toLowerCase();
                li.style.display = optionText.includes(searchTerm)
                    ? ""
                    : "none";
            });
        });

        // Update selected options display
        function updateSelectedOptions() {
            selectedOptions.innerHTML = "";
            optionsList.querySelectorAll("li.selected").forEach((li) => {
                const badge = document.createElement("span");
                badge.className =
                    "bg-blue-100 text-blue-800 text-xs font-medium mr-2 px-2.5 py-0.5 rounded";
                badge.textContent = li.textContent;
                selectedOptions.appendChild(badge);
            });
        }
    });

    const applyFiltersBtn = document.getElementById("apply-filters");
    applyFiltersBtn.addEventListener("click", applyFilters);

    function applyFilters() {
        const selectedFilters = {};
        multiselects.forEach((multiselect) => {
            const filterId = multiselect.previousElementSibling.id;
            const selectedOptions = Array.from(
                multiselect.querySelectorAll(".options-list li.selected")
            ).map((li) => li.dataset.value);
            if (selectedOptions.length > 0) {
                selectedFilters[filterId] = selectedOptions;
            }
        });

        console.log("Selected filters:", selectedFilters);

        // Use fetch to send a request to the server
        fetch("/filtered-quotes", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                "X-CSRF-Token": document
                    .querySelector('meta[name="csrf-token"]')
                    .getAttribute("content"),
            },
            body: JSON.stringify(selectedFilters),
        })
            .then((response) => response.text())
            .then((html) => {
                document.getElementById("quotes-table-body").innerHTML = html;
            })
            .catch((error) => {
                console.error("Error:", error);
            });
    }
});

// Handle new author and book dialogs
document.addEventListener("DOMContentLoaded", function () {
    const addNewAuthorBtn = document.getElementById("add-new-author");
    const addNewBookBtn = document.getElementById("add-new-book");
    const newAuthorDialog = document.getElementById("new-author-dialog");
    const newBookDialog = document.getElementById("new-book-dialog");
    const newAuthorForm = document.getElementById("new-author-form");
    const newBookForm = document.getElementById("new-book-form");
    const selectAuthor = document.getElementById("select-author");
    const selectBook = document.getElementById("select-book");
    const cancelNewAuthorBtn = document.getElementById("cancel-new-author");
    const cancelNewBookBtn = document.getElementById("cancel-new-book");

    if (addNewAuthorBtn) {
        addNewAuthorBtn.addEventListener("click", () => {
            newAuthorDialog.showModal();
        });
    }

    if (addNewBookBtn) {
        addNewBookBtn.addEventListener("click", () => {
            newBookDialog.showModal();
        });
    }

    if (cancelNewAuthorBtn) {
        cancelNewAuthorBtn.addEventListener("click", () => {
            newAuthorDialog.close();
        });
    }

    if (cancelNewBookBtn) {
        cancelNewBookBtn.addEventListener("click", () => {
            newBookDialog.close();
        });
    }

    if (newAuthorForm) {
        newAuthorForm.addEventListener("submit", (e) => {
            e.preventDefault();
            const authorName = document.getElementById("new_author_name").value;
            // Here you would typically send an AJAX request to create the new author
            const option = new Option(authorName, "new_" + Date.now());
            selectAuthor.add(option);
            selectAuthor.value = option.value;
            newAuthorDialog.close();
        });
    }

    if (newBookForm) {
        newBookForm.addEventListener("submit", (e) => {
            e.preventDefault();
            const bookTitle = document.getElementById("new_book_title").value;
            // Here you would typically send an AJAX request to create the new book
            // For now, we'll just add it to the select options
            const option = new Option(bookTitle, "new_" + Date.now());
            selectBook.add(option);
            selectBook.value = option.value;
            newBookDialog.close();
        });
    }
});

// Copy quote to clipboard
document.addEventListener('DOMContentLoaded', function() {
    document.querySelectorAll('#copy-quote-button').forEach(button => {
        button.addEventListener('click', function() {
            const quoteElement = this.closest('.relative').querySelector('#quote-text');
            const quote = quoteElement.textContent.trim();
            const authorElement = this.closest('.relative').querySelector('#quote-author');
            const author = authorElement.textContent.trim();
            navigator.clipboard.writeText(`"${quote}" ${author}`).then(() => {
                const originalTitle = this.getAttribute('title');
                this.setAttribute('title', 'Copied!');
                setTimeout(() => {
                    this.setAttribute('title', originalTitle);
                }, 6000);
            }).catch(err => {
                console.error('Failed to copy text: ', err);
            });
        });
    });
});

// Delete quote
document.addEventListener('DOMContentLoaded', function() {
    const deleteQuoteButton = document.querySelector('#deleteQuoteButton');
    if (deleteQuoteButton) {
        deleteQuoteButton.addEventListener('click', function() {
            if (confirm('Are you sure you want to delete this quote?')) {
                const deleteForm = document.querySelector('#deleteForm');
                if (deleteForm) {
                    deleteForm.submit();
                }
            }
        });
    }
});