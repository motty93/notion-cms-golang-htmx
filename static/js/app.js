// 必要に応じて htmx イベントのカスタム処理を追加します
document.addEventListener("htmx:afterSwap", (e) => {
	console.log("Content updated!", e.detail.target);
});
