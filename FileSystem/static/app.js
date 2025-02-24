document.addEventListener("DOMContentLoaded", function () {
    // Функция для загрузки данных по указанному пути
    function loadFiles(path, order = "desc") {
        fetch(`/api/files?path=${encodeURIComponent(path)}&sort=${order}`)
            .then((response) => response.json())
            .then((data) => {
                const fileList = document.getElementById("file-list");
                fileList.innerHTML = ""; 

                data.forEach(el => {
             
                    const pathDiv = document.createElement("div");
                    pathDiv.className = "grid-item";
                    pathDiv.textContent = el.path;

                    const sizeDiv = document.createElement("div");
                    sizeDiv.className = "grid-item";
                    sizeDiv.textContent = el.size;

                    const typeDiv = document.createElement("div");
                    typeDiv.className = "grid-item";
                    typeDiv.textContent = el.is_dir ? "Директория" : "Файл";

                   
                    if (el.is_dir) {
                        pathDiv.style.cursor = "pointer";
                        pathDiv.addEventListener("click", () => {
                        
                            const newPath = el.path;
                            history.pushState({ path: newPath }, "", `?path=${encodeURIComponent(newPath)}`);
                            currentPath = newPath; 
                            loadFiles(newPath); 
                        });
                    }
                    fileList.appendChild(pathDiv);
                    fileList.appendChild(sizeDiv);
                    fileList.appendChild(typeDiv);
                });
            })
            .catch((error) => console.error("Error fetching files:", error));
    }
   

    
    window.addEventListener("popstate", (event) => {
        const path = event.state?.path || "/home/danil"; 
        loadFiles(path); 
    });

    // Загружаем данные для текущего пути
    const urlParams = new URLSearchParams(window.location.search);
    let currentPath = urlParams.get("path") || "/home/danil"; 
    loadFiles(currentPath);

    const sortAsc = document.querySelector('.sort-asc');
    const sortDesc = document.querySelector('.sort-desc');

    // Обработчик клика для сортировки по возрастанию
    sortAsc.addEventListener('click', () => {
        loadFiles(currentPath, "asc"); 
        history.pushState({ path: currentPath }, "", `?path=${encodeURIComponent(currentPath)}&sort=asc`);
    });
    
    // Обработчик клика для сортировки по убыванию
    sortDesc.addEventListener('click', () => {
        loadFiles(currentPath, "desc"); 
        history.pushState({ path: currentPath }, "", `?path=${encodeURIComponent(currentPath)}&sort=desc`);
    });
    
    const backButton = document.querySelector('.button-back');
    backButton.addEventListener('click', () => {
        const currentPath = new URLSearchParams(window.location.search).get("path") || "/home/danil";
        const parentPath = getParentPath(currentPath); 
    
        if (parentPath !== currentPath) { 
            history.pushState({ path: parentPath }, "", `?path=${encodeURIComponent(parentPath)}`);
            loadFiles(parentPath);
        }
    });

    function getParentPath(path) {
        const segments = path.split('/').filter(segment => segment.length > 0);
        if (segments.length <= 1) {
            return "/"; 
        }
        segments.pop(); 
        return '/' + segments.join('/');
    }
});
