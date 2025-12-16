import { type ReactElement } from "react";
import { Route, Routes } from "react-router-dom";

const App = (): ReactElement => {
  return (
    <>
      <Routes>
        <Route path="/" element={<></>} />
        <Route path="/l/:id" element={<></>} />
      </Routes>
    </>
  );
};

export default App;
