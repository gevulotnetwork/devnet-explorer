let ticks;
let fresh = true;
const time = 1000;
const minDigits = 6

function scrollNumber(counter, digits) {
  counter.querySelectorAll('span[data-value]').forEach((tick, i) => {
    tick.style.transform = `translateY(-${100 * parseInt(digits[i])}%)`;
  })

  counter.style.width = `${digits.length * 5.1}rem`;
}

function addDigit(counter, digit, fresh) {
  const spanList = Array(10)
    .join(0)
    .split(0)
    .map((x, j) => `<span>${j}</span>`)
    .join('')

  counter.insertAdjacentHTML(
    "beforeend",
    `<span style="transform: translateY(-1000%)" data-value="${digit}">
        ${spanList}
      </span>`)

  const firstDigit = counter.lastElementChild

  setTimeout(() => {
    firstDigit.className = "visible";
  }, fresh ? 0 : 2000);
}

function removeDigit(counter) {
  const firstDigit = counter.lastElementChild
  firstDigit.classList.remove("visible");
  setTimeout(() => {
    firstDigit.remove();
  }, 2000);
}

function setup(counter) {
  console.log(counter)
  startNum = 0
  const digits = startNum.toString().split('')

  for (let i = 0; i < minDigits; i++) {
    addDigit(counter, '0', true)
  }

  scrollNumber(counter, ['0'])

  setTimeout(() => scrollNumber(counter, digits), 2000)

  counter.dataset.value = startNum;
}

function rollToNumber(idx, num) {
  el.style.transform = `translateY(-${100 - num * 10}%)`;
}

function update(counter, num) {
  const toDigits = num.toString().split('')
  const fromDigits = counter.dataset.value.toString().split('')
  console.log(fromDigits, toDigits)

  for (let i = fromDigits.length - toDigits.length; i > 0; i--) {
    removeDigit(counter)
  }
  for (let i = toDigits.length - fromDigits.length; i > 0; i--) {
    addDigit(counter, toDigits[i]);
  }

  scrollNumber(counter, toDigits)
  counter.dataset.value = num
}

function fetchData() {
  fetch('/api/v1/stats')
    .then(res => res.json())
    .then(data => {
      for (let key in data) {
        console.log(key, data[key])
        update(document.getElementById(key), data[key])
      }
    })
    .catch(err => console.error(err))
}

function setupCounters() {
  for (const element of document.getElementsByClassName('rolling-number')) {
    setup(element);
  }
}

setupCounters();
setInterval(fetchData, 5000)
