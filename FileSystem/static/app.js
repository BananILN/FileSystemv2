document.addEventListener("DOMContentLoaded", function () {
    // Функция для загрузки данных по указанному пути
    function loadFiles(path, order = "desc") {
        fetch(`/api/files?path=${encodeURIComponent(path)}&sort=${order}`)
            .then((response) => response.json())
            .then((data) => {
                const fileList = document.getElementById("file-list");
                fileList.innerHTML = ""; // Очищаем текущий список

                data.forEach(el => {
                    // Создаем элементы для каждой строки таблицы
                    const pathDiv = document.createElement("div");
                    pathDiv.className = "grid-item";
                    pathDiv.textContent = el.path;

                    const sizeDiv = document.createElement("div");
                    sizeDiv.className = "grid-item";
                    sizeDiv.textContent = el.size;

                    const typeDiv = document.createElement("div");
                    typeDiv.className = "grid-item";
                    typeDiv.textContent = el.is_dir ? "Директория" : "Файл";

                    // Если это директория, добавляем обработчик клика
                    if (el.is_dir) {
                        pathDiv.style.cursor = "pointer";
                        pathDiv.addEventListener("click", () => {
                            // Обновляем URL и загружаем данные для новой директории
                            const newPath = el.path;
                            history.pushState({ path: newPath }, "", `?path=${encodeURIComponent(newPath)}`);
                            currentPath = newPath; // Обновляем currentPath
                            loadFiles(newPath); // Загружаем файлы для новой директории
                        });
                    }
                    fileList.appendChild(pathDiv);
                    fileList.appendChild(sizeDiv);
                    fileList.appendChild(typeDiv);
                });
            })
            .catch((error) => console.error("Error fetching files:", error));
    }

    // Обработчик изменения URL (например, при нажатии кнопки "Назад" в браузере)
    window.addEventListener("popstate", (event) => {
        const path = event.state?.path || "/home/danil"; // Путь по умолчанию
        loadFiles(path); // Загружаем файлы для текущего пути
    });

    // Загружаем данные для текущего пути
    const urlParams = new URLSearchParams(window.location.search);
    let currentPath = urlParams.get("path") || "/home/danil"; // Изменено на let
    loadFiles(currentPath);

    const sortAsc = document.querySelector('.sort-asc');
    const sortDesc = document.querySelector('.sort-desc');

    // Обработчик клика для сортировки по возрастанию
    sortAsc.addEventListener('click', () => {
        loadFiles(currentPath, "asc"); // Передаем текущий путь и сортировку
        history.pushState({ path: currentPath }, "", `?path=${encodeURIComponent(currentPath)}&sort=asc`);
    });
    
    // Обработчик клика для сортировки по убыванию
    sortDesc.addEventListener('click', () => {
        loadFiles(currentPath, "desc"); // Передаем текущий путь и сортировку
        history.pushState({ path: currentPath }, "", `?path=${encodeURIComponent(currentPath)}&sort=desc`);
    });
});