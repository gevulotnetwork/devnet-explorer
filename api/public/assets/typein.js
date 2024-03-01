const charArr = [
  "⫖",
  "⫈",
  "⪸",
  "⪬",
  "⫛",
  "⫏",
  "⫐",
  "⩱",
  "⩸",
  "⩦",
  "⩨",
  "⩢",
  "⩽",
  "⨺",
  "⨻",
  "⩥",
  "⩩",
  "⫒",
  "⫕",
  "⪫",
  "⪭",
  "⫑",
  "⫓",
  "⪷",
  "⪵",
];

let totalDuration = 400; // Total time for the animation in milliseconds

function randomChar() {
  return charArr[Math.floor(Math.random() * charArr.length)];
}
function animateElement(target) {
  let originalText = target.textContent;

  // Remove leading and trailing whitespaces
  originalText = originalText.trim();

  // Replace multiple whitespaces (including tabs and newlines) with a single space
  originalText = originalText.replace(/\s+/g, " ");

  target.textContent = "";

  let wordList = originalText.split(" ");
  let flatCharacters = [];

  wordList.forEach((word, wordIndex) => {
    const wordSpan = document.createElement("span");
    wordSpan.className = "word";

    Array.from(word).forEach((char, charIndex) => {
      let charSpan = document.createElement("span");
      charSpan.classList.add("animating");
      charSpan.style.opacity = "0.0";
      charSpan.textContent = randomChar();
      wordSpan.appendChild(charSpan);

      flatCharacters.push({
        wordIndex,
        charIndex,
        char,
      });
    });

    target.appendChild(wordSpan);

    // Don't add a space after the last word
    if (wordIndex < wordList.length - 1) {
      target.appendChild(document.createTextNode(" "));
    }
  });

  function reveal(index) {
    let wordIndex = flatCharacters[index].wordIndex;
    let charIndex = flatCharacters[index].charIndex;
    let targetChar = flatCharacters[index].char;

    let span = target.querySelectorAll(".word")[wordIndex].children[charIndex];
    span.style.opacity = "1.0";

    let interval = setInterval(() => {
      span.textContent = randomChar();
    }, 50);

    setTimeout(() => {
      clearInterval(interval);
      span.textContent = targetChar;
      span.classList.remove("animating");
      span.classList.add("animating-done");
      span.style.opacity = "1.0";
    }, 1000);
  }

  for (let i = 0; i < flatCharacters.length; i++) {
    let randomTime = Math.random() * totalDuration;
    setTimeout(() => reveal(i), randomTime);
  }
}

const elementsToAnimate = document.querySelectorAll(".animate");
elementsToAnimate.forEach((element) => {
  animateElement(element);
});
