// Add live class to the current page link in the nav
const navLinks = document.querySelectorAll("nav a");
for (var i = 0; i < navLinks.length; i++) {
	const link = navLinks[i]
	if (link.getAttribute('href') == window.location.pathname) {
		link.classList.add("live");
		break;
	}
}

// Get the checkbox element
const mobileNavToggle = document.querySelector("#mobile-nav-toggle");
const mobileNavMenu = document.querySelector("#mobile-nav-menu");
const hamburgerIcon = document.querySelector("#hamburger-icon");
const closeIcon = document.querySelector("#close-icon");
const body = document.querySelector("body");
const mobileNavMenuOverlay = document.querySelector("#mobile-nav-menu-overlay");

// Add event listener to the checkbox
mobileNavToggle.addEventListener("change", function() {
    if (this.checked) {
        mobileNavMenu.classList.remove("invisible");
        mobileNavMenu.classList.remove("left-[100%]");
        mobileNavMenu.classList.add("visible");
        mobileNavMenu.classList.add("left-0");
        hamburgerIcon.classList.add("hidden");
        hamburgerIcon.classList.remove("flex");
        closeIcon.classList.remove("hidden");
        closeIcon.classList.add("flex");
        body.classList.add("fixed");
        mobileNavMenuOverlay.classList.remove("hidden");
        mobileNavMenuOverlay.classList.add("flex");
    } else {
        mobileNavMenu.classList.add("invisible");
        mobileNavMenu.classList.remove("visible");
        mobileNavMenu.classList.remove("left-0");
        mobileNavMenu.classList.add("left-[100%]");
        hamburgerIcon.classList.remove("hidden");
        hamburgerIcon.classList.add("flex");
        closeIcon.classList.add("hidden");
        closeIcon.classList.remove("flex");
        body.classList.remove("fixed");
        mobileNavMenuOverlay.classList.remove("flex");
        mobileNavMenuOverlay.classList.add("hidden");
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