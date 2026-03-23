import IndexView from "../views/IndexView";
import { useEffect } from "react";
import { useNavigate } from "react-router-dom";

export default function Index() {
    const navigate = useNavigate();

    useEffect(() => {
        if (localStorage.getItem("token")) {
            navigate("/dashboard");
        }
    }, []);

    return <IndexView />;
}
